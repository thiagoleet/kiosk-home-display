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
	"github.com/thiagoleet/kiosk-home-display/internal/weather"
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
	printerMon   *printer.Monitor
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

	// Brightness is cosmetic, so a display that refuses the adjustment must not
	// keep the daemon from starting: failing here makes systemd restart-loop
	// over a screen that is otherwise perfectly usable.
	if err := displayManager.SetBrightness(
		cfg.Display.Brightness,
	); err != nil {
		if errors.Is(
			err,
			display.ErrBrightnessUnsupported,
		) {
			log.Printf(
				"[APP] display mode %q ignores brightness: %v",
				cfg.Display.Mode,
				err,
			)
		} else {
			log.Printf(
				"[APP] initial display brightness not applied: %v",
				err,
			)
		}
	}

	idleManager := idle.NewManager(
		bus,
		cfg.Idle.Timeout,
		cfg.Idle.SleepDelay,
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

	// A virtual printer has no queue to watch, so no monitor is created and the
	// simulated job endpoint stays available for development.
	var printerMonitor *printer.Monitor

	if cfg.Printer.Mode == "cups" {
		printerMonitor = printer.NewMonitor(
			printerManager,
			printer.NewCUPSSource(),
			cfg.Printer.PollInterval,
		)
	}

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

	weatherProvider := weather.NewOpenMeteoProvider(
		nil,
		cfg.Weather.OpenMeteoAPIURL,
	)

	weatherService := weather.NewService(
		cfg.Weather.Enabled,
		weatherProvider,
		weather.Location{
			Latitude:  cfg.Weather.Latitude,
			Longitude: cfg.Weather.Longitude,
			Timezone:  cfg.Weather.Timezone,
		},
		cfg.Weather.ForecastDays,
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
		weatherService,
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
		printerMon:   printerMonitor,
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

	if a.printerMon != nil {
		log.Printf(
			"[APP] watching the CUPS queue every %s",
			a.config.Printer.PollInterval,
		)

		a.printerMon.Start()
	} else {
		// Silence here reads as a broken printer rather than a configuration
		// choice: nothing else in the log mentions printing, so a box left on
		// the default mode looks like one whose queue is never seen.
		log.Printf(
			"[APP] printer mode is %q, no print queue is watched. Set PRINTER_MODE=cups to report real jobs.",
			a.config.Printer.Mode,
		)
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
	// The idle timeout itself only reaches the frontend, which raises the
	// screensaver on a screen that is still lit. Powering the display down
	// waits for the second stage of the countdown, so the screensaver is
	// actually visible for a while before the panel goes dark.
	a.bus.Subscribe(
		events.EventIdleSleep,
		func(event events.Event) {
			if err := a.display.Sleep(); err != nil {
				log.Printf(
					"failed to put display to sleep: %v",
					err,
				)
			}
		},
	)

	// Every wake restarts the idle countdown. Without this a display that a
	// notification or the morning schedule turned on inherits whatever was left
	// of the running countdown, and sleeps again moments later.
	a.bus.Subscribe(
		events.EventDisplayStateChanged,
		func(event events.Event) {
			snapshot, ok := event.Data.(display.Snapshot)
			if !ok {
				return
			}

			if snapshot.Power != display.StateOn {
				return
			}

			a.idle.Activity()
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

			// Waking an already lit screen publishes no state change, so a
			// notification arriving during the screensaver window would not
			// reset the countdown through EventDisplayStateChanged and the
			// display would go dark part way through the notification.
			a.idle.Activity()
		},
	)
}

func (a *App) Stop() error {
	log.Println(
		"Stopping Kiosk Home Display daemon...",
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

	if a.printerMon != nil {
		a.printerMon.Stop()
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
		"Kiosk Home Display daemon stopped",
	)

	return nil
}
