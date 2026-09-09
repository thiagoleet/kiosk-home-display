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

	manager := NewManager(bus, 50*time.Millisecond, time.Hour)

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

	manager := NewManager(bus, 100*time.Millisecond, time.Hour)

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

// The countdown has to keep reporting: a single-shot timer would announce
// idleness once per process, so anything that wakes the display afterwards
// would leave it on until the next restart.
func TestIdleTimeoutRepeats(t *testing.T) {
	bus := events.NewBus()

	receivedTimeouts := make(chan struct{}, 8)

	bus.Subscribe(events.EventIdleTimeout, func(event events.Event) {
		receivedTimeouts <- struct{}{}
	})

	manager := NewManager(bus, 40*time.Millisecond, 40*time.Millisecond)

	manager.Start()

	defer manager.Stop()

	for count := 0; count < 2; count++ {
		select {
		case <-receivedTimeouts:
			// Expected.

		case <-time.After(500 * time.Millisecond):
			t.Fatalf(
				"expected idle timeout %d, got none",
				count+1,
			)
		}
	}
}

func TestStopEndsTheTimeouts(t *testing.T) {
	bus := events.NewBus()

	received := make(chan struct{}, 8)

	bus.Subscribe(events.EventIdleTimeout, func(event events.Event) {
		received <- struct{}{}
	})

	manager := NewManager(bus, 40*time.Millisecond, 40*time.Millisecond)

	manager.Start()

	<-received

	manager.Stop()

	// Drain a timeout that may have fired while Stop was being called.
	select {
	case <-received:
	default:
	}

	select {
	case <-received:
		t.Fatal("expected no idle timeout after Stop")

	case <-time.After(150 * time.Millisecond):
		// Expected.
	}
}

// The display sleeps only after the screensaver has had the whole delay on a
// lit screen. Sleeping at the timeout is what hid the screensaver.
func TestSleepFollowsTheTimeoutAfterTheDelay(t *testing.T) {
	bus := events.NewBus()

	received := make(chan events.Type, 8)

	bus.Subscribe(events.EventIdleTimeout, func(event events.Event) {
		received <- event.Type
	})

	bus.Subscribe(events.EventIdleSleep, func(event events.Event) {
		received <- event.Type
	})

	manager := NewManager(
		bus,
		40*time.Millisecond,
		120*time.Millisecond,
	)

	manager.Start()

	defer manager.Stop()

	select {
	case eventType := <-received:
		if eventType != events.EventIdleTimeout {
			t.Fatalf(
				"expected %q first, got %q",
				events.EventIdleTimeout,
				eventType,
			)
		}

	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected idle timeout event")
	}

	select {
	case eventType := <-received:
		t.Fatalf(
			"expected no %q before the sleep delay elapsed",
			eventType,
		)

	case <-time.After(60 * time.Millisecond):
		// Expected: the screensaver is still on a lit screen.
	}

	select {
	case eventType := <-received:
		if eventType != events.EventIdleSleep {
			t.Fatalf(
				"expected %q, got %q",
				events.EventIdleSleep,
				eventType,
			)
		}

	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected idle sleep event")
	}
}

// Activity during the screensaver window has to cancel the pending sleep,
// otherwise the screen goes dark moments after someone touches it.
func TestActivityDuringScreensaverCancelsSleep(t *testing.T) {
	bus := events.NewBus()

	timeouts := make(chan struct{}, 8)
	sleeps := make(chan struct{}, 8)

	bus.Subscribe(events.EventIdleTimeout, func(event events.Event) {
		timeouts <- struct{}{}
	})

	bus.Subscribe(events.EventIdleSleep, func(event events.Event) {
		sleeps <- struct{}{}
	})

	manager := NewManager(
		bus,
		40*time.Millisecond,
		100*time.Millisecond,
	)

	manager.Start()

	defer manager.Stop()

	select {
	case <-timeouts:
		// Expected.

	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected idle timeout event")
	}

	manager.Activity()

	select {
	case <-sleeps:
		t.Fatal("activity should have cancelled the pending sleep")

	case <-time.After(80 * time.Millisecond):
		// Expected.
	}
}
