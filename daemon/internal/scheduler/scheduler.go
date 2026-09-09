package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type Schedule struct {
	On  string
	Off string
}

type Clock func() time.Time

type Scheduler struct {
	bus      *events.Bus
	schedule Schedule
	location *time.Location
	clock    Clock

	// on and off hold the schedule boundaries as minutes since midnight.
	// parseErr keeps the scheduler quiet when either boundary is malformed,
	// rather than guessing which state the display should be in.
	on       int
	off      int
	parseErr error

	// published is the last window the scheduler announced. Comparing the
	// current window against it publishes on transitions only, whichever tick
	// observes them: a minute missed by a stalled process, or skipped by the
	// clock jump a Pi without an RTC makes when NTP lands, still settles the
	// display into the right state.
	published events.Type

	stop chan struct{}
	done chan struct{}

	stopOnce sync.Once
}

func New(
	bus *events.Bus,
	schedule Schedule,
	location *time.Location,
) *Scheduler {
	scheduler := &Scheduler{
		bus:      bus,
		schedule: schedule,
		location: location,
		clock:    time.Now,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}

	on, err := parseMinutes(schedule.On)
	if err != nil {
		scheduler.parseErr = err

		return scheduler
	}

	off, err := parseMinutes(schedule.Off)
	if err != nil {
		scheduler.parseErr = err

		return scheduler
	}

	scheduler.on = on
	scheduler.off = off

	return scheduler
}

func (s *Scheduler) Start() {
	if s.parseErr != nil {
		log.Printf(
			"[SCHEDULER] no schedule applied: %v",
			s.parseErr,
		)
	}

	// Apply the window the current time already sits in before the ticker takes
	// over. Without this a restart inside the off window — and every deploy is
	// one — leaves the display on until the next SCHEDULE_ON minute.
	s.check()

	go s.run()
}

func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stop)
	})

	<-s.done
}

func (s *Scheduler) run() {
	defer close(s.done)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.check()

		case <-s.stop:
			return
		}
	}
}

func (s *Scheduler) check() {
	if s.parseErr != nil {
		return
	}

	desired := events.EventScheduleOff

	if s.displayShouldBeOn(s.clock().In(s.location)) {
		desired = events.EventScheduleOn
	}

	if desired == s.published {
		return
	}

	s.published = desired

	s.bus.Publish(events.Event{
		Type: desired,
	})
}

// displayShouldBeOn reports whether the given time falls inside the on window.
// The window wraps around midnight when the on time is later than the off time,
// which is how an overnight schedule such as 18:00 to 06:00 is expressed.
func (s *Scheduler) displayShouldBeOn(now time.Time) bool {
	current := now.Hour()*60 + now.Minute()

	if s.on == s.off {
		return true
	}

	if s.on < s.off {
		return current >= s.on && current < s.off
	}

	return current >= s.on || current < s.off
}

func parseMinutes(value string) (int, error) {
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid schedule time %q, want HH:MM: %w",
			value,
			err,
		)
	}

	return parsed.Hour()*60 + parsed.Minute(), nil
}
