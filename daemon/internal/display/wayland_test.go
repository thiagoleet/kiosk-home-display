package display

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

const wlrRandrOutput = `HDMI-A-1 "Samsung Electric Company SAMSUNG 0x00000001 (HDMI-A-1)"
  Make: Samsung Electric Company
  Model: SAMSUNG
  Enabled: yes
  Modes:
    1920x1080 px, 60.000000 Hz (preferred, current)
  Position: 0,0
  Scale: 1.000000
HDMI-A-2 "Unknown Unknown (HDMI-A-2)"
  Enabled: no
`

func newStubbedWaylandController() (
	*WaylandController,
	*commandStub,
) {
	stub := newCommandStub()
	stub.outputs["wlr-randr"] = wlrRandrOutput

	controller := NewWaylandController()
	controller.command = stub.run

	return controller, stub
}

func TestWaylandControllerDisablesEveryOutputOnSleep(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	if err := controller.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"wlr-randr",
		"wlr-randr --output HDMI-A-1 --off",
		"wlr-randr --output HDMI-A-2 --off",
	}

	if len(stub.calls) != len(expected) {
		t.Fatalf(
			"expected %d commands, got %v",
			len(expected),
			stub.calls,
		)
	}

	for index, want := range expected {
		call := stub.calls[index]

		got := strings.Join(
			append([]string{call.name}, call.args...),
			" ",
		)

		if got != want {
			t.Fatalf(
				"expected command %q, got %q",
				want,
				got,
			)
		}
	}
}

// Waking has to reach the outputs that sleeping disabled, so a disabled output
// must not be filtered out of the listing.
func TestWaylandControllerEnablesDisabledOutputOnWake(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	if err := controller.Wake(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "wlr-randr --output HDMI-A-2 --on"

	call := stub.calls[len(stub.calls)-1]

	got := strings.Join(
		append([]string{call.name}, call.args...),
		" ",
	)

	if got != want {
		t.Fatalf(
			"expected command %q, got %q",
			want,
			got,
		)
	}
}

func TestWaylandControllerReportsMissingWlrRandr(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	stub.errs["wlr-randr"] = &exec.Error{
		Name: "wlr-randr",
		Err:  exec.ErrNotFound,
	}

	err := controller.Sleep()

	if err == nil {
		t.Fatal("expected an error when wlr-randr is missing")
	}

	if !strings.Contains(err.Error(), "wlr-randr is not installed") {
		t.Fatalf(
			"expected the error to name the package, got %v",
			err,
		)
	}
}

func TestWaylandControllerReportsUnreachableCompositor(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	stub.outputs["wlr-randr"] =
		"failed to connect to display: no such file or directory\n"
	stub.errs["wlr-randr"] = errors.New("exit status 1")

	err := controller.Sleep()

	if err == nil {
		t.Fatal("expected an error when the compositor is unreachable")
	}

	if !strings.Contains(err.Error(), "WAYLAND_DISPLAY") {
		t.Fatalf(
			"expected the error to name the missing environment, got %v",
			err,
		)
	}
}

func TestWaylandControllerReportsBrightnessUnsupported(t *testing.T) {
	controller, _ := newStubbedWaylandController()

	err := controller.SetBrightness(50)

	if !errors.Is(err, ErrBrightnessUnsupported) {
		t.Fatalf(
			"expected ErrBrightnessUnsupported, got %v",
			err,
		)
	}
}

func TestNewControllerBuildsWaylandController(t *testing.T) {
	controller, err := NewController("wayland")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := controller.(*WaylandController); !ok {
		t.Fatalf(
			"expected a WaylandController, got %T",
			controller,
		)
	}
}
