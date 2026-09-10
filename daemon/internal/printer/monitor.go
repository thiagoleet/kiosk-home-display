package printer

import (
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

	// pollLog keeps a queue that cannot be read from filling the journal with
	// one identical line per interval. Only the poll goroutine touches it, and
	// Start polls before that goroutine exists, so the two never overlap.
	pollLog changeLogger
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
// waiting at boot, and the ones already in the history behind them, become the
// baseline instead of arriving as notifications.
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
	queue, err := m.source.Queue()
	if err != nil {
		m.pollLog.Failed(
			"[PRINTER] cannot read the print queue: %v",
			err,
		)

		return
	}

	m.pollLog.Recovered(
		"[PRINTER] print queue readable again",
	)

	m.manager.Sync(queue)
}
