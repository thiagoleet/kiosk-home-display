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

	m.publishNotification(
		events.NewNotification(
			events.NotificationContextPrinter,
			"Impressão iniciada",
			data.Name,
			events.NotificationInfo,
			5000,
		),
	)

	m.publishActivity(
		events.NewActivity(
			events.NotificationContextPrinter,
			events.EventPrinterStarted,
			"Impressão iniciada",
			data.Name,
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

	m.publishNotification(
		events.NewNotification(
			events.NotificationContextPrinter,
			"Impressão concluída",
			data.Name,
			events.NotificationSuccess,
			5000,
		),
	)

	m.publishActivity(
		events.NewActivity(
			events.NotificationContextPrinter,
			events.EventPrinterCompleted,
			"Impressão concluída",
			data.Name,
		),
	)
}

func (m *Manager) publishNotification(
	notification events.Notification,
) {
	events.PublishNotification(
		m.bus,
		notification,
	)
}

func (m *Manager) publishActivity(
	activity events.Activity,
) {
	events.PublishActivity(
		m.bus,
		activity,
	)
}
