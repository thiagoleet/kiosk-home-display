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
