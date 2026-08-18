package scheduler

import (
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

	lastTriggered string

	stop chan struct{}
	done chan struct{}

	stopOnce sync.Once
}

func New(
	bus *events.Bus,
	schedule Schedule,
	location *time.Location,
) *Scheduler {
	return &Scheduler{
		bus:      bus,
		schedule: schedule,
		location: location,
		clock:    time.Now,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
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
	now := s.clock().In(s.location)

	currentTime := now.Format("15:04")

	if currentTime == s.lastTriggered {
		return
	}

	switch currentTime {
	case s.schedule.On:
		s.bus.Publish(events.Event{
			Type: events.EventScheduleOn,
		})

		s.lastTriggered = currentTime

	case s.schedule.Off:
		s.bus.Publish(events.Event{
			Type: events.EventScheduleOff,
		})

		s.lastTriggered = currentTime
	}
}
