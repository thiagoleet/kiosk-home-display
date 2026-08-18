package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/config"
	"github.com/thiagoleet/kiosk-home-display/internal/display"
	"github.com/thiagoleet/kiosk-home-display/internal/events"
	"github.com/thiagoleet/kiosk-home-display/internal/idle"
	"github.com/thiagoleet/kiosk-home-display/internal/scheduler"
)

type App struct {
	config    config.Config
	bus       *events.Bus
	idle      *idle.Manager
	display   *display.Manager
	scheduler *scheduler.Scheduler
}

func New(cfg config.Config) (*App, error) {
	bus := events.NewBus()

	controller, err := display.NewController(cfg.Display.Mode)
	if err != nil {
		return nil, fmt.Errorf(
			"create display controller: %w",
			err,
		)
	}

	displayManager := display.NewManager(controller)

	if err := displayManager.SetBrightness(
		cfg.Display.Brightness,
	); err != nil {
		return nil, fmt.Errorf(
			"set initial display brightness: %w",
			err,
		)
	}

	idleManager := idle.NewManager(
		bus,
		cfg.Idle.Timeout,
	)

	location, err := time.LoadLocation(cfg.Scheduler.Timezone)
	if err != nil {
		return nil, fmt.Errorf(
			"load scheduler timezone: %w",
			err,
		)
	}

	schedulerManager := scheduler.New(
		bus,
		scheduler.Schedule{
			On:  cfg.Scheduler.On,
			Off: cfg.Scheduler.Off,
		},
		location,
	)

	return &App{
		config:    cfg,
		bus:       bus,
		idle:      idleManager,
		display:   displayManager,
		scheduler: schedulerManager,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.registerHandlers()

	if a.config.Idle.Enabled {
		a.idle.Start()
	}

	if a.config.Scheduler.Enabled {
		a.scheduler.Start()
	}

	log.Println("Kiosk Home Display application is running")

	<-ctx.Done()

	log.Println("Shutdown signal received")

	return a.Stop()
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

func (a *App) Stop() error {
	log.Println("Stopping Kiosk Home Display application...")

	if a.config.Scheduler.Enabled {
		a.scheduler.Stop()
	}

	if a.config.Idle.Enabled {
		a.idle.Stop()
	}

	log.Println("Kiosk Home Display application stopped")

	return nil
}
