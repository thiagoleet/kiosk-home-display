package printer

import (
	"errors"
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

func job(id string, name string) PrintJob {
	return PrintJob{ID: id, Name: name}
}

// The first read of the queue is a baseline. Without this, a job stuck in the
// queue announces itself as new every time the daemon restarts.
func TestManagerSyncAdoptsTheFirstQueueSilently(t *testing.T) {
	manager, recorded := newRecordedManager()

	manager.Sync([]PrintJob{job("printer-1", "report.pdf")})

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

	manager.Sync(nil)
	manager.Sync([]PrintJob{job("printer-1", "report.pdf")})

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

	manager.Sync(nil)

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

	queue := []PrintJob{job("printer-1", "report.pdf")}

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

	manager.Sync([]PrintJob{job("printer-1", "first.pdf")})
	manager.Sync([]PrintJob{job("printer-2", "second.pdf")})

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

	manager.Sync([]PrintJob{
		job("printer-1", "first.pdf"),
		job("printer-2", "second.pdf"),
	})

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
