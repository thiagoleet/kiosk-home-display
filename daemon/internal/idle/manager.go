package idle

import (
	"sync"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type Manager struct {
	bus     *events.Bus
	timeout time.Duration

	mu      sync.Mutex
	timer   *time.Timer
	running bool
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

	m.running = true

	m.resetTimer()
}

func (m *Manager) Activity() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.resetTimer()
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.running = false

	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
}

func (m *Manager) resetTimer() {
	if m.timer != nil {
		m.timer.Stop()
	}

	m.timer = time.AfterFunc(m.timeout, m.fire)
}

// fire announces the timeout and immediately arms the next one. A single-shot
// timer would report idleness once per process: anything that wakes the display
// afterwards — a notification, the morning schedule — would leave it on until
// the next restart.
func (m *Manager) fire() {
	m.mu.Lock()

	if !m.running {
		m.mu.Unlock()

		return
	}

	m.resetTimer()

	m.mu.Unlock()

	// Published outside the lock: subscribers report activity back through
	// Activity, which would deadlock on the mutex held here.
	m.bus.Publish(events.Event{
		Type: events.EventIdleTimeout,
	})
}
