package app

import (
	"log"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/display"
	"github.com/thiagoleet/kiosk-home-display/internal/events"
	"github.com/thiagoleet/kiosk-home-display/internal/iddle"
)

type App struct {
	bus     *events.Bus
	iddle   *iddle.Manager
	display *display.Manager
}

func New() *App {
	bus := events.NewBus()
	controller := &display.MockController{}
	displayManager := display.NewManager(controller)
	iddleManager := iddle.NewManager(bus, 5*time.Minute)

	return &App{
		bus:     bus,
		iddle:   iddleManager,
		display: displayManager,
	}

}

func (a *App) Run() error {
	a.registerHandlers()
	a.iddle.Start()

	log.Println("Kiosk Home Display is running")

	select {}

	return nil
}

func (a *App) registerHandlers() {
	a.bus.Subscribe(events.EventIdleTimeout, func(event events.Event) {
		if err := a.display.Sleep(); err != nil {
			log.Printf("failed to put display to sleep: %v", err)
		}
	})
}
