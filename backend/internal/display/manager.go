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

func (m *Manager) Wake() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state == StateOn {
		return nil
	}

	if err := m.controller.Wake(); err != nil {
		return err
	}

	m.state = StateOn

	m.publishStateChanged()

	return nil
}

func (m *Manager) Sleep() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state == StateOff {
		return nil
	}

	if err := m.controller.Sleep(); err != nil {
		return err
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
