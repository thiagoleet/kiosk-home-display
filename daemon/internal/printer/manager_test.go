package printer

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type recordedEvent struct {
	eventType events.Type
	data      events.PrinterEvent
}

func newRecordedManager() (*Manager, *[]recordedEvent) {
	bus := events.NewBus()

	recorded := &[]recordedEvent{}

	collect := func(event events.Event) {
		data, ok := event.Data.(events.PrinterEvent)
		if !ok {
			return
		}

		*recorded = append(*recorded, recordedEvent{
			eventType: event.Type,
			data:      data,
		})
	}

	bus.Subscribe(events.EventPrinterStarted, collect)
	bus.Subscribe(events.EventPrinterCompleted, collect)

	return NewManager(bus), recorded
}

// job builds a job the way a source would, taking the trailing number of the
// identifier as the job number, which is how CUPS identifiers are shaped.
func job(id string, name string) PrintJob {
	number := 0

	if matches := queuedJob.FindStringSubmatch(id); matches != nil {
		number, _ = strconv.Atoi(matches[2])
	}

	return PrintJob{ID: id, Name: name, Number: number}
}

// active is the common case: a queue with nothing finished behind it.
func active(jobs ...PrintJob) Queue {
	return Queue{Active: jobs}
}

// The first read of the queue is a baseline. Without this, a job stuck in the
// queue announces itself as new every time the daemon restarts.
func TestManagerSyncAdoptsTheFirstQueueSilently(t *testing.T) {
	manager, recorded := newRecordedManager()

	manager.Sync(active(job("printer-1", "report.pdf")))

	if len(*recorded) != 0 {
		t.Fatalf(
			"expected no events for the first queue, got %d",
			len(*recorded),
		)
	}

	if manager.Snapshot().State != StatePrinting {
		t.Errorf(
			"expected the snapshot to report printing, got %q",
			manager.Snapshot().State,
		)
	}
}

func TestManagerSyncPublishesStartedAndCompleted(t *testing.T) {
	manager, recorded := newRecordedManager()

	manager.Sync(Queue{})
	manager.Sync(active(job("printer-1", "report.pdf")))

	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}

	started := (*recorded)[0]

	if started.eventType != events.EventPrinterStarted {
		t.Errorf(
			"expected printer.started, got %q",
			started.eventType,
		)
	}

	if started.data.Name != "report.pdf" {
		t.Errorf(
			"expected the document title, got %q",
			started.data.Name,
		)
	}

	if started.data.JobID != "printer-1" {
		t.Errorf(
			"expected the job identifier, got %q",
			started.data.JobID,
		)
	}

	manager.Sync(Queue{})

	if len(*recorded) != 2 {
		t.Fatalf("expected 2 events, got %d", len(*recorded))
	}

	if (*recorded)[1].eventType != events.EventPrinterCompleted {
		t.Errorf(
			"expected printer.completed, got %q",
			(*recorded)[1].eventType,
		)
	}

	snapshot := manager.Snapshot()

	if snapshot.State != StateIdle {
		t.Errorf(
			"expected the snapshot to report idle, got %q",
			snapshot.State,
		)
	}

	if snapshot.Job != nil {
		t.Errorf(
			"expected no job on an empty queue, got %+v",
			snapshot.Job,
		)
	}
}

// A queue that does not change must stay quiet, however often it is polled.
func TestManagerSyncIsQuietForAnUnchangedQueue(t *testing.T) {
	manager, recorded := newRecordedManager()

	queue := active(job("printer-1", "report.pdf"))

	manager.Sync(queue)
	manager.Sync(queue)
	manager.Sync(queue)

	if len(*recorded) != 0 {
		t.Fatalf("expected no events, got %d", len(*recorded))
	}
}

// One poll can observe a job finishing and the next one starting. The finished
// job is reported first, in the order the printer worked through them.
func TestManagerSyncHandlesAHandoverInOnePoll(t *testing.T) {
	manager, recorded := newRecordedManager()

	manager.Sync(active(job("printer-1", "first.pdf")))
	manager.Sync(active(job("printer-2", "second.pdf")))

	if len(*recorded) != 2 {
		t.Fatalf("expected 2 events, got %d", len(*recorded))
	}

	if (*recorded)[0].eventType != events.EventPrinterCompleted ||
		(*recorded)[0].data.JobID != "printer-1" {
		t.Errorf(
			"expected the finished job first, got %+v",
			(*recorded)[0],
		)
	}

	if (*recorded)[1].eventType != events.EventPrinterStarted ||
		(*recorded)[1].data.JobID != "printer-2" {
		t.Errorf(
			"expected the new job second, got %+v",
			(*recorded)[1],
		)
	}
}

// The snapshot follows the head of the queue, which is the job CUPS works
// through, while the jobs behind it wait.
func TestManagerSyncReportsTheHeadOfTheQueue(t *testing.T) {
	manager, _ := newRecordedManager()

	manager.Sync(active(
		job("printer-1", "first.pdf"),
		job("printer-2", "second.pdf"),
	))

	snapshot := manager.Snapshot()

	if snapshot.Job == nil || snapshot.Job.Name != "first.pdf" {
		t.Fatalf(
			"expected the head of the queue, got %+v",
			snapshot.Job,
		)
	}
}

func TestManagerPrintRejectsAnEmptyName(t *testing.T) {
	manager, _ := newRecordedManager()

	if _, err := manager.Print(""); !errors.Is(
		err,
		ErrInvalidJobName,
	) {
		t.Fatalf("expected ErrInvalidJobName, got %v", err)
	}
}

func TestManagerPrintRejectsASecondJob(t *testing.T) {
	manager, _ := newRecordedManager()

	if _, err := manager.Print("report.pdf"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := manager.Print("second.pdf"); !errors.Is(
		err,
		ErrBusy,
	) {
		t.Fatalf("expected ErrBusy, got %v", err)
	}
}

// Simulated jobs would describe a print that never happened, so a watched
// queue turns the endpoint off.
func TestManagerPrintIsRefusedWhileMonitored(t *testing.T) {
	manager, _ := newRecordedManager()

	NewMonitor(manager, &stubSource{}, time.Hour)

	if _, err := manager.Print("report.pdf"); !errors.Is(
		err,
		ErrMonitored,
	) {
		t.Fatalf("expected ErrMonitored, got %v", err)
	}
}

// The whole point of reading the history: a job accepted and finished between
// two polls is never in the queue, and only its number tells the manager it
// ran at all.
func TestManagerSyncReportsAJobMissedBetweenPolls(t *testing.T) {
	manager, recorded := newRecordedManager()

	manager.Sync(Queue{
		Completed: []PrintJob{job("printer-30", "old.pdf")},
	})

	manager.Sync(Queue{
		Completed: []PrintJob{
			job("printer-31", "receipt.pdf"),
			job("printer-30", "old.pdf"),
		},
	})

	if len(*recorded) != 2 {
		t.Fatalf("expected 2 events, got %+v", *recorded)
	}

	if (*recorded)[0].eventType != events.EventPrinterStarted ||
		(*recorded)[0].data.JobID != "printer-31" {
		t.Errorf(
			"expected the missed job to start, got %+v",
			(*recorded)[0],
		)
	}

	if (*recorded)[1].eventType != events.EventPrinterCompleted ||
		(*recorded)[1].data.JobID != "printer-31" {
		t.Errorf(
			"expected the missed job to finish, got %+v",
			(*recorded)[1],
		)
	}
}

// The history a daemon finds at boot describes prints that happened while it
// was down. Replaying it would announce them all over again on every restart.
func TestManagerSyncAdoptsExistingHistorySilently(t *testing.T) {
	manager, recorded := newRecordedManager()

	manager.Sync(Queue{
		Completed: []PrintJob{
			job("printer-33", "third.pdf"),
			job("printer-32", "second.pdf"),
			job("printer-31", "first.pdf"),
		},
	})

	if len(*recorded) != 0 {
		t.Fatalf(
			"expected the history to be adopted silently, got %+v",
			*recorded,
		)
	}
}

// CUPS keeps a finished job in the history for a long time, so it comes back
// on every poll. It must be reported exactly once.
func TestManagerSyncReportsAMissedJobOnlyOnce(t *testing.T) {
	manager, recorded := newRecordedManager()

	manager.Sync(Queue{})

	history := Queue{
		Completed: []PrintJob{job("printer-31", "receipt.pdf")},
	}

	manager.Sync(history)
	manager.Sync(history)
	manager.Sync(history)

	if len(*recorded) != 2 {
		t.Fatalf(
			"expected the missed job reported once, got %+v",
			*recorded,
		)
	}
}

// A job watched through the queue lands in the history too. The difference
// between the queues already reported it, so the history must not repeat it.
func TestManagerSyncDoesNotRepeatAWatchedJob(t *testing.T) {
	manager, recorded := newRecordedManager()

	manager.Sync(Queue{})
	manager.Sync(active(job("printer-31", "report.pdf")))

	manager.Sync(Queue{
		Completed: []PrintJob{job("printer-31", "report.pdf")},
	})

	if len(*recorded) != 2 {
		t.Fatalf("expected 2 events, got %+v", *recorded)
	}

	if (*recorded)[0].eventType != events.EventPrinterStarted ||
		(*recorded)[1].eventType != events.EventPrinterCompleted {
		t.Errorf(
			"expected one start then one completion, got %+v",
			*recorded,
		)
	}
}

// Several short jobs can pass between two polls. They are reported in the
// order the printer worked through them, which is the order of their numbers.
func TestManagerSyncOrdersMissedJobsOldestFirst(t *testing.T) {
	manager, recorded := newRecordedManager()

	manager.Sync(Queue{})

	// lpstat lists the history newest first, so the input is reversed.
	manager.Sync(Queue{
		Completed: []PrintJob{
			job("printer-33", "third.pdf"),
			job("printer-32", "second.pdf"),
			job("printer-31", "first.pdf"),
		},
	})

	expected := []string{
		"printer-31",
		"printer-31",
		"printer-32",
		"printer-32",
		"printer-33",
		"printer-33",
	}

	if len(*recorded) != len(expected) {
		t.Fatalf(
			"expected %d events, got %+v",
			len(expected),
			*recorded,
		)
	}

	for index, id := range expected {
		if (*recorded)[index].data.JobID != id {
			t.Errorf(
				"event %d: expected %q, got %q",
				index,
				id,
				(*recorded)[index].data.JobID,
			)
		}
	}
}

// A restarted cupsd counts from the beginning again, so the numbers no longer
// relate to the ones already seen. Replaying the new history as missed jobs
// would announce prints that already happened.
func TestManagerSyncReSeedsWhenTheCounterRestarts(t *testing.T) {
	manager, recorded := newRecordedManager()

	manager.Sync(Queue{
		Completed: []PrintJob{job("printer-33", "third.pdf")},
	})

	manager.Sync(Queue{
		Completed: []PrintJob{job("printer-1", "after-restart.pdf")},
	})

	if len(*recorded) != 0 {
		t.Fatalf(
			"expected the restarted counter to re-seed, got %+v",
			*recorded,
		)
	}

	// Counting resumes from the new baseline.
	manager.Sync(Queue{
		Completed: []PrintJob{
			job("printer-2", "next.pdf"),
			job("printer-1", "after-restart.pdf"),
		},
	})

	if len(*recorded) != 2 {
		t.Fatalf(
			"expected the next job reported, got %+v",
			*recorded,
		)
	}

	if (*recorded)[0].data.JobID != "printer-2" {
		t.Errorf(
			"expected the job after the restart, got %+v",
			(*recorded)[0],
		)
	}
}

// A history that empties out must not read as a counter restart: it carries no
// number to compare, so the watermark has to survive it.
func TestManagerSyncKeepsTheWatermarkAcrossAnEmptyQueue(t *testing.T) {
	manager, recorded := newRecordedManager()

	manager.Sync(Queue{
		Completed: []PrintJob{job("printer-31", "old.pdf")},
	})

	manager.Sync(Queue{})

	manager.Sync(Queue{
		Completed: []PrintJob{job("printer-31", "old.pdf")},
	})

	if len(*recorded) != 0 {
		t.Fatalf(
			"expected no events for a job already adopted, got %+v",
			*recorded,
		)
	}
}
