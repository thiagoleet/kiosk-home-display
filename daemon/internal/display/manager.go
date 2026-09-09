package display

import (
	"fmt"
	"sync"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type Manager struct {
	controller Controller
	bus        *events.Bus

	mu         sync.RWMutex
	state      State
	brightness int
}

type Snapshot struct {
	Power      State `json:"power"`
	Brightness int   `json:"brightness"`
}

func NewManager(
	controller Controller,
	bus *events.Bus,
) *Manager {
	return &Manager{
		controller: controller,
		bus:        bus,
		state:      StateOn,
		brightness: 100,
	}
}

func (m *Manager) State() State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.state
}

// Wake drives the controller even when the manager already believes the display
// is on. That belief is only ever a guess: it starts as StateOn at every
// restart, whatever the screen is actually doing, and anything else on the box
// can power the screen down behind the daemon's back. Skipping the controller
// on a stale belief is how the API comes to answer 200 while the screen stays
// dark. Every controller is idempotent, so re-asserting costs one cheap
// command, and the state change is still published only on a transition.
func (m *Manager) Wake() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.controller.Wake(); err != nil {
		return err
	}

	if m.state == StateOn {
		return nil
	}

	m.state = StateOn

	m.publishStateChanged()

	return nil
}

// Sleep re-asserts the controller for the same reason Wake does.
func (m *Manager) Sleep() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.controller.Sleep(); err != nil {
		return err
	}

	if m.state == StateOff {
		return nil
	}

	m.state = StateOff

	m.publishStateChanged()

	return nil
}

func (m *Manager) SetBrightness(level int) error {
	if level < 0 || level > 100 {
		return fmt.Errorf(
			"brightness must be between 0 and 100",
		)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.controller.SetBrightness(level); err != nil {
		return err
	}

	m.brightness = level

	m.publishStateChanged()

	return nil
}

func (m *Manager) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return Snapshot{
		Power:      m.state,
		Brightness: m.brightness,
	}
}

func (m *Manager) publishStateChanged() {
	snapshot := Snapshot{
		Power:      m.state,
		Brightness: m.brightness,
	}

	m.bus.Publish(events.Event{
		Type: events.EventDisplayStateChanged,
		Data: snapshot,
	})
}
