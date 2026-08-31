package activity

import (
	"context"
	"log"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
	"github.com/thiagoleet/kiosk-home-display/internal/i18n"
)

type Manager struct {
	bus        *events.Bus
	repository Repository
	texts      i18n.Catalog
	lifeSpan   time.Duration
}

func NewManager(
	bus *events.Bus,
	repository Repository,
	texts i18n.Catalog,
	lifeSpan time.Duration,
) *Manager {
	return &Manager{
		bus:        bus,
		repository: repository,
		texts:      texts,
		lifeSpan:   lifeSpan,
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

	if err := m.repository.DeleteOlderThan(
		context.Background(),
		time.Now().Add(-m.lifeSpan),
	); err != nil {
		log.Printf(
			"failed to delete old activities: %v",
			err,
		)
	}
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
		m.texts.Text(i18n.KeyPrinterStarted),
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
		m.texts.Text(i18n.KeyPrinterCompleted),
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
