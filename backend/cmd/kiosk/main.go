package main

import (
	"log"
	"time"
)

func main() {
	log.Println("Kiosk Home Display backend starting…")

	for {
		log.Println("Backend is running")

		time.Sleep(10 * time.Second)
	}
}
