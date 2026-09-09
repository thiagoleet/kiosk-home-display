package printer

import (
	"log"
	"sync"
	"time"
)

// Monitor polls a print system and hands what it sees to the manager, which
// turns the difference between two polls into printer events.
//
// Polling is the only mechanism available without a CUPS subscription: cupsd
// can push notifications, but only to a notifier registered in its own
// configuration, which a user-level daemon cannot install.
type Monitor struct {
	manager  *Manager
	source   Source
	interval time.Duration

	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once

	// lastErr keeps a queue that cannot be read from filling the journal with
	// one identical line per interval. Only the poll goroutine touches it, and
	// Start polls before that goroutine exists, so the two never overlap.
	lastErr string
}

func NewMonitor(
	manager *Manager,
	source Source,
	interval time.Duration,
) *Monitor {
	// A watched queue is the authority on what is printing, so the manager
	// stops accepting simulated jobs from the HTTP API.
	manager.setMonitored()

	return &Monitor{
		manager:  manager,
		source:   source,
		interval: interval,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// Start reads the queue once before the ticker takes over, so the jobs already
// waiting at boot become the baseline instead of arriving as notifications.
func (m *Monitor) Start() {
	m.poll()

	go m.run()
}

func (m *Monitor) Stop() {
	m.stopOnce.Do(func() {
		close(m.stop)
	})

	<-m.done
}

func (m *Monitor) run() {
	defer close(m.done)

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.poll()

		case <-m.stop:
			return
		}
	}
}

func (m *Monitor) poll() {
	jobs, err := m.source.ActiveJobs()
	if err != nil {
		m.reportError(err)

		return
	}

	m.clearError()

	m.manager.Sync(jobs)
}

// reportError logs a failing poll once, and again only when the failure
// changes. A stopped cupsd is otherwise reported every interval forever.
func (m *Monitor) reportError(err error) {
	message := err.Error()

	if message == m.lastErr {
		return
	}

	m.lastErr = message

	log.Printf(
		"[PRINTER] cannot read the print queue: %v",
		err,
	)
}

func (m *Monitor) clearError() {
	if m.lastErr == "" {
		return
	}

	m.lastErr = ""

	log.Println("[PRINTER] print queue readable again")
}
