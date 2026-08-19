package notification

import (
	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type Manager struct {
	bus *events.Bus
}

func NewManager(
	bus *events.Bus,
) *Manager {
	return &Manager{
		bus: bus,
	}
}

func (m *Manager) Start() {
	m.bus.Subscribe(
		events.EventPrinterStarted,
		m.handlePrinterStarted,
	)

	m.bus.Subscribe(
		events.EventPrinterCompleted,
		m.handlePrinterCompleted,
	)
}

func (m *Manager) handlePrinterStarted(
	event events.Event,
) {
	data, ok := event.Data.(events.PrinterEvent)
	if !ok {
		return
	}

	m.publish(
		events.NewNotification(
			events.NotificationContextPrinter,
			"Impressão iniciada",
			data.Name,
			events.NotificationInfo,
			5000,
		),
	)
}

func (m *Manager) handlePrinterCompleted(
	event events.Event,
) {
	data, ok := event.Data.(events.PrinterEvent)
	if !ok {
		return
	}

	m.publish(
		events.NewNotification(
			events.NotificationContextPrinter,
			"Impressão concluída",
			data.Name,
			events.NotificationSuccess,
			5000,
		),
	)
}

func (m *Manager) publish(
	notification events.Notification,
) {
	events.PublishNotification(
		m.bus,
		notification,
	)
}
