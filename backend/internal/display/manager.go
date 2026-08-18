package display

type Manager struct {
	controller Controller
	state      State
}

func NewManager(controller Controller) *Manager {
	return &Manager{
		controller: controller,
		state:      StateOn,
	}
}

func (m *Manager) Wake() error {
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
	if m.state == StateOff {
		return nil
	}

	if err := m.controller.Sleep(); err != nil {
		return err
	}

	m.state = StateOff
	return nil
}

func (m *Manager) Dim() error {
	if m.state == StateDimmed {
		return nil
	}

	if err := m.controller.Dim(); err != nil {
		return err
	}

	m.state = StateDimmed
	return nil
}
