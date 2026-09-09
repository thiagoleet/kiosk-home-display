package printer

import "errors"

var (
	// ErrInvalidJobName is returned when a simulated job carries no name.
	ErrInvalidJobName = errors.New(
		"print job name cannot be empty",
	)

	// ErrBusy is returned while another job holds the printer.
	ErrBusy = errors.New(
		"printer is already printing",
	)

	// ErrMonitored is returned when the daemon watches a real CUPS queue.
	// Simulating a job there would publish events for a print that never
	// happened, and the next poll would contradict them.
	ErrMonitored = errors.New(
		"printer jobs come from CUPS, simulation is disabled",
	)
)
