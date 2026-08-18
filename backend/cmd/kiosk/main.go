package main

import (
	"log"

	"github.com/thiagoleet/kiosk-home-display/internal/app"
)

func main() {
	application := app.New()

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
