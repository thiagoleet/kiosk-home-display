package iddle

import (
	"testing"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

func TestManagerPublishesSleepAfterTimeout(t *testing.T) {
	bus := events.NewBus()

	eventReceived := make(chan events.Event, 1)

	bus.Subscribe(events.EventDisplaySleep, func(event events.Event) {
		eventReceived <- event
	})

	manager := NewManager(bus, 50*time.Millisecond)

	manager.Start()

	select {
	case <-eventReceived:
		// Expected
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected display sleep event")
	}

	manager.Stop()
}

func TestActivityResetsTimer(t *testing.T) {
	bus := events.NewBus()

	eventReceived := make(chan events.Event, 1)

	bus.Subscribe(events.EventDisplaySleep, func(event events.Event) {
		eventReceived <- event
	})

	manager := NewManager(bus, 100*time.Millisecond)

	manager.Start()

	manager.Activity() // Reset the timer

	select {
	case <-eventReceived:
		t.Fatal("display should not sleep yet")
	case <-time.After(70 * time.Millisecond):
		// Expected
	}

	select {
	case <-eventReceived:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected display sleep event")

	}

	manager.Stop()
}
