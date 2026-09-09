package display

import (
	"testing"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

func TestManagerStartsWithDisplayOn(t *testing.T) {
	bus := events.NewBus()
	controller := NewVirtualController()
	manager := NewManager(controller, bus)

	if manager.State() != StateOn {
		t.Fatalf(
			"expected display state %q, got %q",
			StateOn,
			manager.State(),
		)
	}

	if controller.State() != StateOn {
		t.Fatalf(
			"expected controller state %q, got %q",
			StateOn,
			controller.State(),
		)
	}
}

func TestManagerSleepsDisplay(t *testing.T) {
	bus := events.NewBus()
	controller := NewVirtualController()
	manager := NewManager(controller, bus)

	err := manager.Sleep()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if manager.State() != StateOff {
		t.Fatalf(
			"expected manager state %q, got %q",
			StateOff,
			manager.State(),
		)
	}

	if controller.State() != StateOff {
		t.Fatalf(
			"expected controller state %q, got %q",
			StateOff,
			controller.State(),
		)
	}
}

func TestManagerWakesDisplay(t *testing.T) {
	controller := NewVirtualController()
	bus := events.NewBus()
	manager := NewManager(controller, bus)

	if err := manager.Sleep(); err != nil {
		t.Fatalf("failed to sleep display: %v", err)
	}

	if err := manager.Wake(); err != nil {
		t.Fatalf("failed to wake display: %v", err)
	}

	if manager.State() != StateOn {
		t.Fatalf(
			"expected manager state %q, got %q",
			StateOn,
			manager.State(),
		)
	}

	if controller.State() != StateOn {
		t.Fatalf(
			"expected controller state %q, got %q",
			StateOn,
			controller.State(),
		)
	}
}

func TestManagerDoesNotWakeAlreadyAwakeDisplay(t *testing.T) {
	controller := NewVirtualController()
	bus := events.NewBus()
	manager := NewManager(controller, bus)

	if err := manager.Wake(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if manager.State() != StateOn {
		t.Fatalf("expected display to remain on")
	}
}

func TestManagerSetsBrightness(t *testing.T) {
	controller := NewVirtualController()
	bus := events.NewBus()
	manager := NewManager(controller, bus)

	err := manager.SetBrightness(75)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if controller.Brightness() != 75 {
		t.Fatalf(
			"expected brightness 75, got %d",
			controller.Brightness(),
		)
	}
}

func TestManagerRejectsInvalidBrightness(t *testing.T) {
	bus := events.NewBus()
	controller := NewVirtualController()
	manager := NewManager(controller, bus)

	tests := []int{
		-1,
		101,
	}

	for _, brightness := range tests {
		err := manager.SetBrightness(brightness)

		if err == nil {
			t.Fatalf(
				"expected error for brightness %d",
				brightness,
			)
		}
	}
}

// countingController records how often the manager drives it.
type countingController struct {
	wakes  int
	sleeps int
}

func (c *countingController) Wake() error {
	c.wakes++

	return nil
}

func (c *countingController) Sleep() error {
	c.sleeps++

	return nil
}

func (c *countingController) SetBrightness(level int) error {
	return nil
}

// The manager's state starts as StateOn at every restart, whatever the screen
// is doing. Trusting it would answer the caller with success while the screen
// stays dark, so both transitions re-assert the controller.
func TestManagerReassertsTheControllerOnStaleState(t *testing.T) {
	bus := events.NewBus()
	controller := &countingController{}
	manager := NewManager(controller, bus)

	if err := manager.Wake(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if controller.wakes != 1 {
		t.Fatalf(
			"expected the controller to be woken once, got %d",
			controller.wakes,
		)
	}

	if err := manager.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := manager.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if controller.sleeps != 2 {
		t.Fatalf(
			"expected the controller to be slept twice, got %d",
			controller.sleeps,
		)
	}
}

// Re-asserting must not turn into an event storm: only a real transition is
// published.
func TestManagerPublishesOnlyOnTransitions(t *testing.T) {
	bus := events.NewBus()
	controller := &countingController{}
	manager := NewManager(controller, bus)

	published := 0

	bus.Subscribe(
		events.EventDisplayStateChanged,
		func(event events.Event) {
			published++
		},
	)

	if err := manager.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := manager.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if published != 1 {
		t.Fatalf(
			"expected 1 published state change, got %d",
			published,
		)
	}
}
