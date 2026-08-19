package notification

import (
	"github.com/thiagoleet/kiosk-home-display/internal/events"
	"github.com/thiagoleet/kiosk-home-display/internal/i18n"
)

type Manager struct {
	bus   *events.Bus
	texts i18n.Catalog
}

func NewManager(
	bus *events.Bus,
	texts i18n.Catalog,
) *Manager {
	return &Manager{
		bus:   bus,
		texts: texts,
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
			m.texts.Text(i18n.KeyPrinterStarted),
			data.Name,
			events.NotificationInfo,
			5000,
		),
	)

	m.publishActivity(
		events.NewActivity(
			events.NotificationContextPrinter,
			events.EventPrinterStarted,
			m.texts.Text(i18n.KeyPrinterStarted),
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
			m.texts.Text(i18n.KeyPrinterCompleted),
			data.Name,
			events.NotificationSuccess,
			5000,
		),
	)

	m.publishActivity(
		events.NewActivity(
			events.NotificationContextPrinter,
			events.EventPrinterCompleted,
			m.texts.Text(i18n.KeyPrinterCompleted),
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
