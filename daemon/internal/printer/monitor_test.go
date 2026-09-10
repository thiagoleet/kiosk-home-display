package printer

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

// stubSource answers successive polls with successive queues, repeating the
// last one once it runs out, and reports every poll on a channel so a test can
// wait for the loop instead of sleeping for it.
type stubSource struct {
	mu     sync.Mutex
	queues []Queue
	calls  int
	err    error

	polled chan struct{}
}

func newStubSource(queues ...Queue) *stubSource {
	return &stubSource{
		queues: queues,
		polled: make(chan struct{}, 64),
	}
}

func (s *stubSource) Queue() (Queue, error) {
	s.mu.Lock()

	index := s.calls

	s.calls++

	err := s.err

	var queue Queue

	if len(s.queues) > 0 {
		if index >= len(s.queues) {
			index = len(s.queues) - 1
		}

		queue = s.queues[index]
	}

	s.mu.Unlock()

	select {
	case s.polled <- struct{}{}:

	default:
	}

	if err != nil {
		return Queue{}, err
	}

	return queue, nil
}

func (s *stubSource) waitForPolls(
	t *testing.T,
	count int,
) {
	t.Helper()

	for i := 0; i < count; i++ {
		select {
		case <-s.polled:

		case <-time.After(2 * time.Second):
			t.Fatalf(
				"timed out waiting for poll %d of %d",
				i+1,
				count,
			)
		}
	}
}

func TestMonitorPublishesQueueChanges(t *testing.T) {
	manager, recorded := newRecordedManager()

	source := newStubSource(
		active(job("printer-1", "report.pdf")),
		active(job("printer-1", "report.pdf")),
		active(
			job("printer-1", "report.pdf"),
			job("printer-2", "list.txt"),
		),
		active(job("printer-2", "list.txt")),
		Queue{},
	)

	monitor := NewMonitor(
		manager,
		source,
		5*time.Millisecond,
	)

	monitor.Start()

	source.waitForPolls(t, 5)

	monitor.Stop()

	// The job in the queue at startup is the baseline, so the events describe
	// the second job arriving and both jobs leaving.
	expected := []recordedEvent{
		{
			eventType: events.EventPrinterStarted,
			data: events.PrinterEvent{
				JobID: "printer-2",
				Name:  "list.txt",
			},
		},
		{
			eventType: events.EventPrinterCompleted,
			data: events.PrinterEvent{
				JobID: "printer-1",
				Name:  "report.pdf",
			},
		},
		{
			eventType: events.EventPrinterCompleted,
			data: events.PrinterEvent{
				JobID: "printer-2",
				Name:  "list.txt",
			},
		},
	}

	if len(*recorded) != len(expected) {
		t.Fatalf(
			"expected %d events, got %d: %+v",
			len(expected),
			len(*recorded),
			*recorded,
		)
	}

	for index, event := range expected {
		if (*recorded)[index] != event {
			t.Errorf(
				"event %d: expected %+v, got %+v",
				index,
				event,
				(*recorded)[index],
			)
		}
	}
}

// A queue that cannot be read must not look like an empty one: publishing
// completions for a stopped cupsd would clear jobs that are still queued.
func TestMonitorIgnoresAFailingQueue(t *testing.T) {
	manager, recorded := newRecordedManager()

	source := newStubSource(
		active(job("printer-1", "report.pdf")),
	)

	monitor := NewMonitor(
		manager,
		source,
		5*time.Millisecond,
	)

	source.err = errors.New("lpstat: Unable to connect to server")

	monitor.Start()

	source.waitForPolls(t, 3)

	monitor.Stop()

	if len(*recorded) != 0 {
		t.Fatalf(
			"expected no events, got %+v",
			*recorded,
		)
	}

	if manager.Snapshot().State != StateIdle {
		t.Errorf(
			"expected the snapshot to stay idle, got %q",
			manager.Snapshot().State,
		)
	}
}

func TestMonitorStopEndsTheLoop(t *testing.T) {
	manager, _ := newRecordedManager()

	source := newStubSource(Queue{})

	monitor := NewMonitor(
		manager,
		source,
		5*time.Millisecond,
	)

	monitor.Start()
	source.waitForPolls(t, 2)
	monitor.Stop()

	source.mu.Lock()
	before := source.calls
	source.mu.Unlock()

	time.Sleep(50 * time.Millisecond)

	source.mu.Lock()
	after := source.calls
	source.mu.Unlock()

	if after != before {
		t.Errorf(
			"expected polling to stop, went from %d to %d calls",
			before,
			after,
		)
	}
}

// End to end through the poll loop: a job that never appears in the queue is
// still reported, from the history it leaves behind.
func TestMonitorPublishesAJobThatOnlyReachesTheHistory(t *testing.T) {
	manager, recorded := newRecordedManager()

	source := newStubSource(
		Queue{
			Completed: []PrintJob{job("printer-30", "old.pdf")},
		},
		Queue{
			Completed: []PrintJob{
				job("printer-31", "receipt.pdf"),
				job("printer-30", "old.pdf"),
			},
		},
	)

	monitor := NewMonitor(
		manager,
		source,
		5*time.Millisecond,
	)

	monitor.Start()

	source.waitForPolls(t, 2)

	monitor.Stop()

	expected := []recordedEvent{
		{
			eventType: events.EventPrinterStarted,
			data: events.PrinterEvent{
				JobID: "printer-31",
				Name:  "receipt.pdf",
			},
		},
		{
			eventType: events.EventPrinterCompleted,
			data: events.PrinterEvent{
				JobID: "printer-31",
				Name:  "receipt.pdf",
			},
		},
	}

	if len(*recorded) != len(expected) {
		t.Fatalf(
			"expected %d events, got %+v",
			len(expected),
			*recorded,
		)
	}

	for index, event := range expected {
		if (*recorded)[index] != event {
			t.Errorf(
				"event %d: expected %+v, got %+v",
				index,
				event,
				(*recorded)[index],
			)
		}
	}
}
