package display

import (
	"errors"
	"os"
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

type waylandCommandStub struct {
	*commandStub

	env []string
}

func (s *waylandCommandStub) run(
	env []string,
	name string,
	args ...string,
) (string, error) {
	s.env = env

	return s.commandStub.run(name, args...)
}

// The stub describes a session the way a Pi does: nothing in the environment,
// a runtime directory named after the service user, one compositor socket.
func newStubbedWaylandController() (
	*WaylandController,
	*waylandCommandStub,
) {
	stub := &waylandCommandStub{
		commandStub: newCommandStub(),
	}

	stub.outputs["wlr-randr"] = wlrRandrOutput

	controller := NewWaylandController()
	controller.command = stub.run
	controller.uid = 1000
	// The fallback path by default: most of these tests cover wlr-randr.
	controller.lookPath = func(name string) (string, error) {
		return "", &exec.Error{
			Name: name,
			Err:  exec.ErrNotFound,
		}
	}
	controller.lookupEnv = func(string) (string, bool) {
		return "", false
	}
	controller.sockets = func(directory string) ([]string, error) {
		return []string{
			directory + "/wayland-1",
		}, nil
	}

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

// A system service inherits no session variables, so the controller has to
// derive both from the user it runs as.
func TestWaylandControllerDerivesTheSessionFromTheServiceUser(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	if err := controller.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"XDG_RUNTIME_DIR=/run/user/1000",
		"WAYLAND_DISPLAY=wayland-1",
	}

	for _, want := range expected {
		found := false

		for _, entry := range stub.env {
			if entry == want {
				found = true
			}
		}

		if !found {
			t.Fatalf(
				"expected %q in the command environment, got %v",
				want,
				stub.env,
			)
		}
	}
}

func TestWaylandControllerKeepsSessionVariablesFromTheEnvironment(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	controller.lookupEnv = func(name string) (string, bool) {
		switch name {
		case "XDG_RUNTIME_DIR":
			return "/run/user/1001", true

		case "WAYLAND_DISPLAY":
			return "wayland-0", true
		}

		return "", false
	}

	controller.sockets = func(string) ([]string, error) {
		t.Fatal("expected no socket lookup when the environment is set")

		return nil, nil
	}

	if err := controller.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"XDG_RUNTIME_DIR=/run/user/1001",
		"WAYLAND_DISPLAY=wayland-0",
	}

	for _, want := range expected {
		found := false

		for _, entry := range stub.env {
			if entry == want {
				found = true
			}
		}

		if !found {
			t.Fatalf(
				"expected %q in the command environment, got %v",
				want,
				stub.env,
			)
		}
	}
}

func TestWaylandControllerReportsMissingSession(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	controller.sockets = func(string) ([]string, error) {
		return nil, nil
	}

	err := controller.Sleep()

	if err == nil {
		t.Fatal("expected an error when no socket exists")
	}

	if !strings.Contains(err.Error(), "/run/user/1000") {
		t.Fatalf(
			"expected the error to name the runtime directory, got %v",
			err,
		)
	}

	if len(stub.calls) != 0 {
		t.Fatalf(
			"expected no commands, got %v",
			stub.calls,
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

	if !strings.Contains(err.Error(), "wayland-1") {
		t.Fatalf(
			"expected the error to name the socket it tried, got %v",
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

func TestWaylandSocketsSkipsLockFiles(t *testing.T) {
	directory := t.TempDir()

	for _, name := range []string{
		"wayland-0",
		"wayland-0.lock",
	} {
		if err := writeEmptyFile(directory + "/" + name); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	sockets, err := waylandSockets(directory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sockets) != 1 ||
		!strings.HasSuffix(sockets[0], "wayland-0") {
		t.Fatalf(
			"expected only the socket, got %v",
			sockets,
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

func writeEmptyFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	return file.Close()
}

// wlopm powers the sink without touching the output layout, so it is the path
// to take whenever it is installed.
func TestWaylandControllerPrefersWlopm(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	controller.lookPath = func(name string) (string, error) {
		if name == "wlopm" {
			return "/usr/bin/wlopm", nil
		}

		return "", &exec.Error{
			Name: name,
			Err:  exec.ErrNotFound,
		}
	}

	if err := controller.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stub.calls) != 1 {
		t.Fatalf(
			"expected one command, got %v",
			stub.calls,
		)
	}

	want := "wlopm --off *"

	got := strings.Join(
		append(
			[]string{stub.calls[0].name},
			stub.calls[0].args...,
		),
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

func TestWaylandControllerWakesWithWlopm(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	controller.lookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}

	if err := controller.Wake(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "wlopm --on *"

	got := strings.Join(
		append(
			[]string{stub.calls[0].name},
			stub.calls[0].args...,
		),
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

// A compositor that refuses to re-enable a disabled output leaves the screen
// dark, so the error has to point at the tool that avoids the problem.
func TestWaylandControllerSuggestsWlopmWhenConfigurationIsRefused(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	stub.outputs["wlr-randr --output HDMI-A-1 --on"] =
		"failed to apply configuration\n"
	stub.errs["wlr-randr --output HDMI-A-1 --on"] =
		errors.New("exit status 1")

	err := controller.Wake()

	if err == nil {
		t.Fatal("expected an error when the configuration is refused")
	}

	if !strings.Contains(err.Error(), "wlopm") {
		t.Fatalf(
			"expected the error to suggest wlopm, got %v",
			err,
		)
	}
}
