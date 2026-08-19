package printer

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type Manager struct {
	bus *events.Bus

	mu       sync.RWMutex
	snapshot Snapshot
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

func (m *Manager) Print(
	name string,
) (PrintJob, error) {
	if name == "" {
		return PrintJob{}, fmt.Errorf(
			"print job name cannot be empty",
		)
	}

	m.mu.Lock()

	if m.snapshot.State == StatePrinting {
		m.mu.Unlock()

		return PrintJob{}, fmt.Errorf(
			"printer is already printing",
		)
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

	m.bus.Publish(events.Event{
		Type: events.EventPrinterStarted,
		Data: events.PrinterEvent{
			JobID: job.ID,
			Name:  job.Name,
		},
	})

	go m.simulatePrint(job)

	return job, nil
}

func (m *Manager) simulatePrint(
	job PrintJob,
) {
	time.Sleep(5 * time.Second)

	m.mu.Lock()

	m.snapshot = Snapshot{
		State: StateCompleted,
		Job:   &job,
	}

	m.mu.Unlock()

	m.bus.Publish(events.Event{
		Type: events.EventPrinterCompleted,
		Data: events.PrinterEvent{
			JobID: job.ID,
			Name:  job.Name,
		},
	})

	time.Sleep(1 * time.Second)

	m.mu.Lock()

	m.snapshot = Snapshot{
		State: StateIdle,
	}

	m.mu.Unlock()
}
