package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	nethttp "net/http"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/activity"
	"github.com/thiagoleet/kiosk-home-display/internal/config"
	"github.com/thiagoleet/kiosk-home-display/internal/database"
	"github.com/thiagoleet/kiosk-home-display/internal/display"
	"github.com/thiagoleet/kiosk-home-display/internal/events"
	"github.com/thiagoleet/kiosk-home-display/internal/http"
	"github.com/thiagoleet/kiosk-home-display/internal/i18n"
	"github.com/thiagoleet/kiosk-home-display/internal/idle"
	"github.com/thiagoleet/kiosk-home-display/internal/notification"
	"github.com/thiagoleet/kiosk-home-display/internal/printer"
	"github.com/thiagoleet/kiosk-home-display/internal/scheduler"
	"github.com/thiagoleet/kiosk-home-display/internal/state"
	"github.com/thiagoleet/kiosk-home-display/internal/websocket"
)

type App struct {
	config       config.Config
	bus          *events.Bus
	db           *database.Database
	idle         *idle.Manager
	display      *display.Manager
	scheduler    *scheduler.Scheduler
	printer      *printer.Manager
	notification *notification.Manager
	activity     *activity.Manager
	websocket    *websocket.Server
	httpServer   *http.Server
}

func New(cfg config.Config) (*App, error) {
	bus := events.NewBus()

	databaseConfig := database.DefaultConfig()

	db, err := database.Open(databaseConfig)
	if err != nil {
		return nil, fmt.Errorf(
			"open database: %w",
			err,
		)
	}

	if err := database.Migrate(db.DB); err != nil {
		db.Close()

		return nil, fmt.Errorf(
			"migrate database: %w",
			err,
		)
	}

	texts, err := i18n.LoadPtBR()
	if err != nil {
		db.Close()

		return nil, fmt.Errorf(
			"load translations: %w",
			err,
		)
	}

	controller, err := display.NewController(
		cfg.Display.Mode,
	)
	if err != nil {
		db.Close()

		return nil, fmt.Errorf(
			"create display controller: %w",
			err,
		)
	}

	displayManager := display.NewManager(
		controller,
		bus,
	)

	if err := displayManager.SetBrightness(
		cfg.Display.Brightness,
	); err != nil {
		db.Close()

		return nil, fmt.Errorf(
			"set initial display brightness: %w",
			err,
		)
	}

	idleManager := idle.NewManager(
		bus,
		cfg.Idle.Timeout,
	)

	location, err := time.LoadLocation(
		cfg.Scheduler.Timezone,
	)
	if err != nil {
		db.Close()

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

	stateManager := state.NewManager(
		displayManager,
	)

	printerManager := printer.NewManager(
		bus,
	)

	notificationManager := notification.NewManager(
		bus,
		texts,
	)

	activityRepository :=
		activity.NewSQLiteRepository(db.DB)

	activityManager := activity.NewManager(
		bus,
		activityRepository,
		texts,
		cfg.Activity.LifeSpan,
	)

	websocketServer := websocket.NewServer(
		bus,
		stateManager,
		cfg.HTTP.AllowedOrigins,
	)

	httpServer := http.NewServer(
		cfg.HTTP.Host,
		cfg.HTTP.Port,
		bus,
		websocketServer,
		displayManager,
		printerManager,
		activityRepository,
		texts,
		cfg.HTTP.AllowedOrigins,
	)

	return &App{
		config: cfg,
		bus:    bus,
		db:     db,

		idle:         idleManager,
		display:      displayManager,
		scheduler:    schedulerManager,
		printer:      printerManager,
		notification: notificationManager,
		activity:     activityManager,

		websocket:  websocketServer,
		httpServer: httpServer,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.registerHandlers()

	a.websocket.Start()
	a.notification.Start()
	a.activity.Start()

	if a.config.Idle.Enabled {
		a.idle.Start()
	}

	if a.config.Scheduler.Enabled {
		a.scheduler.Start()
	}

	go func() {
		if err := a.httpServer.Start(); err != nil {
			if !errors.Is(
				err,
				nethttp.ErrServerClosed,
			) {
				log.Printf(
					"HTTP server error: %v",
					err,
				)
			}
		}
	}()

	log.Println(
		"Kiosk Home Display application is running",
	)

	<-ctx.Done()

	log.Println("Shutdown signal received")

	return a.Stop()
}

func (a *App) registerHandlers() {
	a.bus.Subscribe(
		events.EventIdleTimeout,
		func(event events.Event) {
			if err := a.display.Sleep(); err != nil {
				log.Printf(
					"failed to put display to sleep: %v",
					err,
				)
			}
		},
	)

	a.bus.Subscribe(
		events.EventScheduleOn,
		func(event events.Event) {
			if err := a.display.Wake(); err != nil {
				log.Printf(
					"failed to wake display: %v",
					err,
				)
			}
		},
	)

	a.bus.Subscribe(
		events.EventScheduleOff,
		func(event events.Event) {
			if err := a.display.Sleep(); err != nil {
				log.Printf(
					"failed to put display to sleep: %v",
					err,
				)
			}
		},
	)

	a.bus.Subscribe(
		events.EventNotification,
		func(event events.Event) {
			if err := a.display.Wake(); err != nil {
				log.Printf(
					"failed to wake display for notification: %v",
					err,
				)
			}
		},
	)
}

func (a *App) Stop() error {
	log.Println(
		"Stopping Kiosk Home Display backend...",
	)

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := a.httpServer.Stop(
		shutdownCtx,
	); err != nil {
		log.Printf(
			"failed to stop HTTP server: %v",
			err,
		)
	}

	if a.config.Scheduler.Enabled {
		a.scheduler.Stop()
	}

	if a.config.Idle.Enabled {
		a.idle.Stop()
	}

	if err := a.db.Close(); err != nil {
		log.Printf(
			"failed to close database: %v",
			err,
		)
	}

	log.Println(
		"Kiosk Home Display backend stopped",
	)

	return nil
}
