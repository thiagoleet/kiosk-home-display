package app

import (
	"log"
	"time"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Run() error {
	log.Println("Kiosk Home Display backend is running")

	for {
		time.Sleep(10 * time.Second)
	}

	return nil
}
