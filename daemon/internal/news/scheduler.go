package news

import (
	"context"
	"sync"
	"time"
)

// Scheduler refreshes the feeds in the background. The first refresh runs as
// soon as it starts but on its own goroutine, so a portal that is down or a
// slow connection never delays the kiosk coming up; the slideshow shows what
// was persisted last time in the meantime.
type Scheduler struct {
	service  *Service
	interval time.Duration

	cancel   context.CancelFunc
	done     chan struct{}
	stopOnce sync.Once
}

func NewScheduler(
	service *Service,
	interval time.Duration,
) *Scheduler {
	return &Scheduler{
		service:  service,
		interval: interval,
		done:     make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	s.cancel = cancel

	go s.run(ctx)
}

// Stop cancels a refresh in flight rather than waiting for its feeds to time
// out, so shutdown stays within the daemon's grace period.
func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		if s.cancel == nil {
			close(s.done)

			return
		}

		s.cancel()
	})

	<-s.done
}

func (s *Scheduler) run(ctx context.Context) {
	defer close(s.done)

	s.service.Refresh(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.service.Refresh(ctx)

		case <-ctx.Done():
			return
		}
	}
}
