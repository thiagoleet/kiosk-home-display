package app

import (
	"log"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/display"
	"github.com/thiagoleet/kiosk-home-display/internal/events"
	"github.com/thiagoleet/kiosk-home-display/internal/idle"
	"github.com/thiagoleet/kiosk-home-display/internal/scheduler"
)

type App struct {
	bus       *events.Bus
	idle      *idle.Manager
	display   *display.Manager
	scheduler *scheduler.Scheduler
}

func New() *App {
	bus := events.NewBus()

	controller := &display.MockController{}

	displayManager := display.NewManager(controller)

	idleManager := idle.NewManager(
		bus,
		5*time.Minute,
	)

	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		panic(err)
	}

	scheduler := scheduler.New(
		bus,
		scheduler.Schedule{
			On:  "07:00",
			Off: "23:00",
		},
		location,
	)

	return &App{
		bus:       bus,
		idle:      idleManager,
		display:   displayManager,
		scheduler: scheduler,
	}
}

func (a *App) Run() error {
	a.registerHandlers()

	a.idle.Start()
	a.scheduler.Start()

	log.Println("Kiosk Home Display backend is running")

	select {}

	return nil
}

func (a *App) registerHandlers() {
	a.bus.Subscribe(events.EventIdleTimeout, func(event events.Event) {
		if err := a.display.Sleep(); err != nil {
			log.Printf("failed to put display to sleep: %v", err)
		}
	})

	a.bus.Subscribe(events.EventScheduleOn, func(event events.Event) {
		if err := a.display.Wake(); err != nil {
			log.Printf("failed to wake display: %v", err)
		}
	})

	a.bus.Subscribe(events.EventScheduleOff, func(event events.Event) {
		if err := a.display.Sleep(); err != nil {
			log.Printf("failed to put display to sleep: %v", err)
		}
	})
}
