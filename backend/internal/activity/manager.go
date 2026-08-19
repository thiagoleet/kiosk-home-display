package activity

import (
	"context"
	"log"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type Manager struct {
	bus        *events.Bus
	repository Repository
}

func NewManager(
	bus *events.Bus,
	repository Repository,
) *Manager {
	return &Manager{
		bus:        bus,
		repository: repository,
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

	activity := events.NewActivity(
		events.NotificationContextPrinter,
		events.EventPrinterStarted,
		"Impressão iniciada",
		data.Name,
	)

	m.publish(activity)
}

func (m *Manager) handlePrinterCompleted(
	event events.Event,
) {
	data, ok := event.Data.(events.PrinterEvent)
	if !ok {
		return
	}

	activity := events.NewActivity(
		events.NotificationContextPrinter,
		events.EventPrinterCompleted,
		"Impressão concluída",
		data.Name,
	)

	m.publish(activity)
}

func (m *Manager) publish(
	activity events.Activity,
) {
	if err := m.repository.Create(
		context.Background(),
		activity,
	); err != nil {
		log.Printf(
			"failed to persist activity: %v",
			err,
		)

		return
	}

	events.PublishActivity(
		m.bus,
		activity,
	)
}
