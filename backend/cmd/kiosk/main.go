package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/thiagoleet/kiosk-home-display/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	application := app.New()

	if err := application.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
