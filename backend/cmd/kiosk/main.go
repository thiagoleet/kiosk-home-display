package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/thiagoleet/kiosk-home-display/internal/app"
	"github.com/thiagoleet/kiosk-home-display/internal/config"
	"github.com/thiagoleet/kiosk-home-display/internal/database"
)

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("failed to load .env file: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	databaseConfig := database.DefaultConfig()

	db, err := database.Open(databaseConfig)
	if err != nil {
		log.Fatalf(
			"failed to open database: %v",
			err,
		)
	}

	defer db.Close()

	if err := database.Migrate(db.DB); err != nil {
		log.Fatalf(
			"failed to migrate database: %v",
			err,
		)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
