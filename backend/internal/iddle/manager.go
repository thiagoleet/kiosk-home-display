package iddle

import (
	"sync"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type Manager struct {
	bus     *events.Bus
	timeout time.Duration

	mu    sync.Mutex
	timer *time.Timer
}

func NewManager(bus *events.Bus, timeout time.Duration) *Manager {
	return &Manager{
		bus:     bus,
		timeout: timeout,
	}
}

func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.resetTimer()
}

func (m *Manager) Activity() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.resetTimer()
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
}

func (m *Manager) resetTimer() {
	if m.timer != nil {
		m.timer.Stop()
	}

	m.timer = time.AfterFunc(m.timeout, func() {
		m.bus.Publish(events.Event{
			Type: events.EventDisplaySleep,
		})
	})
}
