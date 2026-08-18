package app

import (
	"log"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type App struct {
	bus *events.Bus
}

func New() *App {
	return &App{
		bus: events.NewBus(),
	}
}

func (a *App) Run() error {
	a.registerHandlers()

	log.Println("Kiosk Home Display backend is running")

	for {
		time.Sleep(10 * time.Second)
	}

	return nil
}

func (a *App) registerHandlers() {
	a.bus.Subscribe(events.EventNotification, func(event events.Event) {
		log.Printf("Notification received: %v", event.Data)
	})
}
