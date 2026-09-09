package printer

import (
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

	// seeded marks the first sync as done. The jobs found by that first read
	// are adopted silently: a job left in the queue would otherwise announce
	// itself as new on every restart of the daemon.
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

// Sync publishes the difference between the queue it is given and the queue the
// previous call saw: a job that appeared has started, and a job that is gone
// has left the queue.
//
// CUPS reports no reason for a job leaving, so a cancelled or aborted job is
// indistinguishable from one that printed and both are published as completed.
func (m *Manager) Sync(jobs []PrintJob) {
	m.mu.Lock()

	previous := m.active

	m.active = jobs
	m.snapshot = snapshotOf(jobs)

	seeded := m.seeded

	m.seeded = true

	m.mu.Unlock()

	if !seeded {
		return
	}

	// Completions first: within one poll a finished job precedes the next one
	// picking up the printer.
	for _, job := range previous {
		if !containsJob(jobs, job.ID) {
			m.publish(events.EventPrinterCompleted, job)
		}
	}

	for _, job := range jobs {
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
