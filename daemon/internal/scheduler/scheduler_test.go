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

// A restart inside the off window must not wait for the next SCHEDULE_ON
// minute: the display is on and nothing else would ever turn it off.
func TestSchedulerAppliesTheCurrentWindowOnStart(t *testing.T) {
	bus := events.NewBus()

	eventReceived := make(chan events.Event, 1)

	bus.Subscribe(events.EventScheduleOff, func(event events.Event) {
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
			23,
			30,
			0,
			0,
			location,
		)
	}

	scheduler.Start()

	defer scheduler.Stop()

	select {
	case <-eventReceived:
		// Expected.
	default:
		t.Fatal("expected schedule.off event")
	}
}

// The boundary minute can pass unobserved: a stalled process, or the clock jump
// a Pi without an RTC makes when NTP lands. The next tick has to settle the
// display anyway.
func TestSchedulerPublishesOffAfterMissedBoundary(t *testing.T) {
	bus := events.NewBus()

	eventReceived := make(chan events.Event, 1)

	bus.Subscribe(events.EventScheduleOff, func(event events.Event) {
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

	now := time.Date(
		2026,
		time.August,
		18,
		22,
		59,
		0,
		0,
		location,
	)

	scheduler.clock = func() time.Time {
		return now
	}

	scheduler.check()

	now = now.Add(46 * time.Minute)

	scheduler.check()

	select {
	case <-eventReceived:
		// Expected.
	default:
		t.Fatal("expected schedule.off event")
	}
}

func TestSchedulerHandlesOvernightWindow(t *testing.T) {
	bus := events.NewBus()

	eventReceived := make(chan events.Event, 1)

	bus.Subscribe(events.EventScheduleOn, func(event events.Event) {
		eventReceived <- event
	})

	location := time.UTC

	scheduler := New(
		bus,
		Schedule{
			On:  "18:00",
			Off: "06:00",
		},
		location,
	)

	scheduler.clock = func() time.Time {
		return time.Date(
			2026,
			time.August,
			18,
			2,
			0,
			0,
			0,
			location,
		)
	}

	scheduler.check()

	select {
	case <-eventReceived:
		// Expected.
	default:
		t.Fatal("expected schedule.on event")
	}
}

func TestSchedulerStaysQuietOnMalformedSchedule(t *testing.T) {
	bus := events.NewBus()

	eventCount := 0

	handler := func(event events.Event) {
		eventCount++
	}

	bus.Subscribe(events.EventScheduleOn, handler)
	bus.Subscribe(events.EventScheduleOff, handler)

	scheduler := New(
		bus,
		Schedule{
			On:  "7am",
			Off: "23:00",
		},
		time.UTC,
	)

	scheduler.check()

	if eventCount != 0 {
		t.Fatalf(
			"expected no events, got %d",
			eventCount,
		)
	}
}
