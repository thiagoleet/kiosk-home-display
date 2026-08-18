package state

import (
	"github.com/thiagoleet/kiosk-home-display/internal/display"
)

type Manager struct {
	display *display.Manager
}

func NewManager(displayManager *display.Manager) *Manager {
	return &Manager{
		display: displayManager,
	}
}

func (m *Manager) Snapshot() State {
	displaySnapshot := m.display.Snapshot()

	return State{
		Display: DisplayState{
			Power:      displaySnapshot.Power,
			Brightness: displaySnapshot.Brightness,
		},
	}
}
