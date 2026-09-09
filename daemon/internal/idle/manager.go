package idle

import (
	"sync"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

// Manager reports idleness in two stages. The first announces the timeout so
// the frontend can raise the screensaver while the panel is still lit; the
// second, sleepDelay later, asks for the display to be powered off. Sleeping
// at the first stage is what made the screensaver pointless: the frontend
// switched to it on a screen that had already gone dark.
type Manager struct {
	bus        *events.Bus
	timeout    time.Duration
	sleepDelay time.Duration

	mu      sync.Mutex
	timer   *time.Timer
	running bool
}

func NewManager(
	bus *events.Bus,
	timeout time.Duration,
	sleepDelay time.Duration,
) *Manager {
	return &Manager{
		bus:        bus,
		timeout:    timeout,
		sleepDelay: sleepDelay,
	}
}

func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.running = true

	m.resetTimer()
}

// Activity restarts the countdown from the first stage, which also cancels a
// sleep the screensaver window has already scheduled.
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
	m.armTimer(m.timeout, m.fireTimeout)
}

func (m *Manager) armTimer(
	delay time.Duration,
	fire func(),
) {
	if m.timer != nil {
		m.timer.Stop()
	}

	m.timer = time.AfterFunc(delay, fire)
}

// fireTimeout announces idleness and leaves the display alone: the screensaver
// needs a lit panel to be seen. The countdown moves to its second stage, which
// powers the screen down once the screensaver has had sleepDelay on screen.
func (m *Manager) fireTimeout() {
	m.mu.Lock()

	if !m.running {
		m.mu.Unlock()

		return
	}

	m.armTimer(m.sleepDelay, m.fireSleep)

	m.mu.Unlock()

	// Published outside the lock: subscribers report activity back through
	// Activity, which would deadlock on the mutex held here.
	m.bus.Publish(events.Event{
		Type: events.EventIdleTimeout,
	})
}

// fireSleep asks for the display to sleep and arms the first stage again. A
// single-shot countdown would report idleness once per process: anything that
// wakes the display afterwards — a notification, the morning schedule — would
// leave it on until the next restart.
func (m *Manager) fireSleep() {
	m.mu.Lock()

	if !m.running {
		m.mu.Unlock()

		return
	}

	m.resetTimer()

	m.mu.Unlock()

	m.bus.Publish(events.Event{
		Type: events.EventIdleSleep,
	})
}
