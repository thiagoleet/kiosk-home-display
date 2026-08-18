package idle

import (
	"testing"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

func TestManagerPublishesIdleTimeoutAfterTimeout(t *testing.T) {
	bus := events.NewBus()

	eventReceived := make(chan events.Event, 1)

	bus.Subscribe(events.EventIdleTimeout, func(event events.Event) {
		eventReceived <- event
	})

	manager := NewManager(bus, 50*time.Millisecond)

	manager.Start()

	defer manager.Stop()

	select {
	case event := <-eventReceived:
		if event.Type != events.EventIdleTimeout {
			t.Fatalf(
				"expected event type %q, got %q",
				events.EventIdleTimeout,
				event.Type,
			)
		}

	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected idle timeout event")
	}
}

func TestActivityResetsTimer(t *testing.T) {
	bus := events.NewBus()

	eventReceived := make(chan events.Event, 1)

	bus.Subscribe(events.EventIdleTimeout, func(event events.Event) {
		eventReceived <- event
	})

	manager := NewManager(bus, 100*time.Millisecond)

	manager.Start()

	defer manager.Stop()

	// Allow part of the timeout to elapse.
	time.Sleep(50 * time.Millisecond)

	// Activity should reset the timer.
	manager.Activity()

	// The original timer would have expired around now,
	// but the reset timer should still be active.
	select {
	case <-eventReceived:
		t.Fatal("idle timeout should have been reset")
	case <-time.After(70 * time.Millisecond):
		// Expected.
	}

	// The new timer should eventually expire.
	select {
	case event := <-eventReceived:
		if event.Type != events.EventIdleTimeout {
			t.Fatalf(
				"expected event type %q, got %q",
				events.EventIdleTimeout,
				event.Type,
			)
		}

	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected idle timeout event")
	}
}
