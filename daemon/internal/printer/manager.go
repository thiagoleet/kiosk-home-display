package printer

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type Manager struct {
	bus *events.Bus

	mu       sync.RWMutex
	snapshot Snapshot

	// active is the queue the last sync observed, in the order the jobs are
	// worked through, so a poll can tell a new job from one that has left.
	active []PrintJob

	// watermark is the highest job number any sync has seen, in the queue or
	// in the history behind it. A finished job numbered above it is one that
	// ran entirely between two polls, which is the only trace such a job
	// leaves.
	watermark int

	// seeded marks the first sync as done. The jobs found by that first read
	// are adopted silently: a job left in the queue, or one sitting in the
	// history, would otherwise announce itself as new on every restart of the
	// daemon.
	seeded bool

	// monitored records that a real queue is being watched, which rules out
	// simulating jobs through the HTTP API.
	monitored bool
}

func NewManager(bus *events.Bus) *Manager {
	return &Manager{
		bus: bus,
		snapshot: Snapshot{
			State: StateIdle,
		},
	}
}

func (m *Manager) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.snapshot
}

// Sync publishes what changed between the queue it is given and the queue the
// previous call saw: a job that appeared has started, and a job that is gone
// has left the queue.
//
// The difference alone is blind to a job that was accepted and finished inside
// one poll interval, which never appears in the queue at all. Those are
// recovered from the history, by their job number: CUPS counts jobs up across
// the whole server, so anything numbered above everything already seen is a
// job that ran unobserved.
//
// CUPS reports no reason for a job leaving, so a cancelled or aborted job is
// indistinguishable from one that printed and both are published as completed.
func (m *Manager) Sync(queue Queue) {
	highest := highestNumber(queue)

	m.mu.Lock()

	previous := m.active
	seeded := m.seeded
	watermark := m.watermark

	// A counter that has gone backwards means cupsd restarted or had its job
	// history cleared, so the numbers no longer line up with the ones already
	// seen. CUPS drops the oldest jobs first, which leaves the highest one
	// alone, so trimming does not look like this. The queue is adopted afresh
	// rather than replayed.
	restarted := highest > 0 && highest < watermark

	m.active = queue.Active
	m.snapshot = snapshotOf(queue.Active)
	m.seeded = true

	if highest > watermark || restarted {
		m.watermark = highest
	}

	m.mu.Unlock()

	if !seeded || restarted {
		return
	}

	// Completions first: within one poll a finished job precedes the next one
	// picking up the printer.
	for _, job := range previous {
		if !containsJob(queue.Active, job.ID) {
			m.publish(events.EventPrinterCompleted, job)
		}
	}

	// Then the jobs that were never watched, which both started and finished
	// while the daemon was between polls.
	for _, job := range missedJobs(queue, previous, watermark) {
		m.publish(events.EventPrinterStarted, job)
		m.publish(events.EventPrinterCompleted, job)
	}

	for _, job := range queue.Active {
		if !containsJob(previous, job.ID) {
			m.publish(events.EventPrinterStarted, job)
		}
	}
}

// Print simulates a job for development and for exercising the frontend. It is
// refused once a real queue is monitored, where the events would describe a
// print that never happened.
func (m *Manager) Print(
	name string,
) (PrintJob, error) {
	if name == "" {
		return PrintJob{}, ErrInvalidJobName
	}

	m.mu.Lock()

	if m.monitored {
		m.mu.Unlock()

		return PrintJob{}, ErrMonitored
	}

	if m.snapshot.State == StatePrinting {
		m.mu.Unlock()

		return PrintJob{}, ErrBusy
	}

	job := PrintJob{
		ID:   uuid.NewString(),
		Name: name,
	}

	m.snapshot = Snapshot{
		State: StatePrinting,
		Job:   &job,
	}

	m.mu.Unlock()

	m.publish(events.EventPrinterStarted, job)

	go m.simulatePrint(job)

	return job, nil
}

func (m *Manager) setMonitored() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.monitored = true
}

func (m *Manager) simulatePrint(
	job PrintJob,
) {
	time.Sleep(10 * time.Second)

	m.mu.Lock()

	m.snapshot = Snapshot{
		State: StateCompleted,
		Job:   &job,
	}

	m.mu.Unlock()

	m.publish(events.EventPrinterCompleted, job)

	time.Sleep(1 * time.Second)

	m.mu.Lock()

	m.snapshot = Snapshot{
		State: StateIdle,
	}

	m.mu.Unlock()
}

func (m *Manager) publish(
	eventType events.Type,
	job PrintJob,
) {
	m.bus.Publish(events.Event{
		Type: eventType,
		Data: events.PrinterEvent{
			JobID: job.ID,
			Name:  job.Name,
		},
	})
}

// snapshotOf describes a queue. The job held by the snapshot is the one at the
// head of the queue, which is the job CUPS is working through.
func snapshotOf(jobs []PrintJob) Snapshot {
	if len(jobs) == 0 {
		return Snapshot{
			State: StateIdle,
		}
	}

	job := jobs[0]

	return Snapshot{
		State: StatePrinting,
		Job:   &job,
	}
}

// highestNumber is the largest job number the queue carries, counting both
// the jobs still to run and the ones already finished.
func highestNumber(queue Queue) int {
	highest := 0

	for _, job := range queue.Active {
		if job.Number > highest {
			highest = job.Number
		}
	}

	for _, job := range queue.Completed {
		if job.Number > highest {
			highest = job.Number
		}
	}

	return highest
}

// missedJobs are the finished jobs no sync ever saw running: numbered above
// everything observed so far, and absent from both the previous queue and the
// current one. They come back oldest first, the order the printer worked
// through them.
func missedJobs(
	queue Queue,
	previous []PrintJob,
	watermark int,
) []PrintJob {
	var missed []PrintJob

	for _, job := range queue.Completed {
		if job.Number <= watermark {
			continue
		}

		// Already accounted for: either it is being reported as completed
		// just above, or it is still running and will be when it leaves.
		if containsJob(previous, job.ID) ||
			containsJob(queue.Active, job.ID) {
			continue
		}

		missed = append(missed, job)
	}

	sort.Slice(missed, func(i int, j int) bool {
		return missed[i].Number < missed[j].Number
	})

	return missed
}

func containsJob(
	jobs []PrintJob,
	id string,
) bool {
	for _, job := range jobs {
		if job.ID == id {
			return true
		}
	}

	return false
}
