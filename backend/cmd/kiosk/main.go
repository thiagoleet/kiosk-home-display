package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/thiagoleet/kiosk-home-display/internal/app"
	"github.com/thiagoleet/kiosk-home-display/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
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
