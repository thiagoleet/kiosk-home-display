package display

import (
	"fmt"
	"sync"
)

type Manager struct {
	controller Controller

	mu    sync.RWMutex
	state State
}

func NewManager(controller Controller) *Manager {
	return &Manager{
		controller: controller,
		state:      StateOn,
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

	return m.controller.SetBrightness(level)
}
