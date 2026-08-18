package scheduler

import (
	"testing"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

func TestSchedulerPublishesScheduleOn(t *testing.T) {
	bus := events.NewBus()

	eventReceived := make(chan events.Event, 1)

	bus.Subscribe(events.EventScheduleOn, func(event events.Event) {
		eventReceived <- event
	})

	location := time.UTC

	scheduler := New(
		bus,
		Schedule{
			On:  "07:00",
			Off: "23:00",
		},
		location,
	)

	scheduler.clock = func() time.Time {
		return time.Date(
			2026,
			time.August,
			18,
			7,
			0,
			0,
			0,
			location,
		)
	}

	scheduler.check()

	select {
	case <-eventReceived:
		// Expected
	default:
		t.Fatal("expected schedule.on event")
	}
}

func TestSchedulerDoesNotTriggerSameScheduleTwice(t *testing.T) {
	bus := events.NewBus()

	eventCount := 0

	bus.Subscribe(events.EventScheduleOn, func(event events.Event) {
		eventCount++
	})

	location := time.UTC

	scheduler := New(
		bus,
		Schedule{
			On:  "07:00",
			Off: "23:00",
		},
		location,
	)

	scheduler.clock = func() time.Time {
		return time.Date(
			2026,
			time.August,
			18,
			7,
			0,
			0,
			0,
			location,
		)
	}

	scheduler.check()
	scheduler.check()

	if eventCount != 1 {
		t.Fatalf(
			"expected 1 event, got %d",
			eventCount,
		)
	}
}
