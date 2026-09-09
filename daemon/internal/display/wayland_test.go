package display

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// HDMI-A-2 is disabled, which is the state the wlr-randr fallback leaves an
// output in after a sleep.
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
  Modes:
    1280x720 px, 60.000000 Hz (preferred)
`

type waylandCommandStub struct {
	*commandStub

	// installed stands in for PATH, so a test can describe a host with or
	// without wlopm.
	installed map[string]bool

	// environment stands in for the process environment.
	environment map[string]string

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

func (s *waylandCommandStub) commands() []string {
	var commands []string

	for _, call := range s.calls {
		commands = append(commands, strings.Join(
			append([]string{call.name}, call.args...),
			" ",
		))
	}

	return commands
}

// The stub describes a session the way a Pi does: nothing in the environment,
// a runtime directory named after the service user, one compositor socket. Only
// wlr-randr is installed, which is the host the fallback path exists for;
// installTool adds wlopm where a test needs it.
func newStubbedWaylandController() (
	*WaylandController,
	*waylandCommandStub,
) {
	stub := &waylandCommandStub{
		commandStub: newCommandStub(),
		installed: map[string]bool{
			"wlr-randr": true,
		},
		environment: map[string]string{},
	}

	stub.outputs["wlr-randr"] = wlrRandrOutput

	controller := NewWaylandController()
	controller.command = stub.run
	controller.uid = 1000
	controller.lookPath = func(name string) (string, error) {
		if stub.installed[name] {
			return "/usr/bin/" + name, nil
		}

		return "", &exec.Error{
			Name: name,
			Err:  exec.ErrNotFound,
		}
	}
	controller.lookupEnv = func(name string) (string, bool) {
		value, ok := stub.environment[name]

		return value, ok
	}
	controller.sockets = func(directory string) ([]string, error) {
		return []string{
			directory + "/wayland-1",
		}, nil
	}

	return controller, stub
}

func assertCommands(
	t *testing.T,
	stub *waylandCommandStub,
	expected ...string,
) {
	t.Helper()

	got := stub.commands()

	if len(got) != len(expected) {
		t.Fatalf(
			"expected commands %v, got %v",
			expected,
			got,
		)
	}

	for index, want := range expected {
		if got[index] != want {
			t.Fatalf(
				"expected command %q, got %q",
				want,
				got[index],
			)
		}
	}
}

// wlopm powers the sink and leaves the output configured, so it is the path to
// take whenever it is installed.
func TestWaylandControllerPrefersWlopmOnSleep(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	stub.installed["wlopm"] = true

	if err := controller.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertCommands(t, stub, "wlopm --off *")
}

// Nothing can power-manage a disabled output, so a screen the fallback switched
// off has to be re-enabled before wlopm is asked to power it on.
func TestWaylandControllerReEnablesOutputsBeforePoweringOn(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	stub.installed["wlopm"] = true

	if err := controller.Wake(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertCommands(
		t,
		stub,
		"wlr-randr",
		"wlr-randr --output HDMI-A-2 --on --preferred",
		"wlopm --on *",
	)
}

func TestWaylandControllerDisablesEnabledOutputsWhenAllowed(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	stub.environment["WAYLAND_ALLOW_OUTPUT_DISABLE"] = "true"

	if err := controller.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertCommands(
		t,
		stub,
		"wlr-randr",
		"wlr-randr --output HDMI-A-1 --off",
	)
}

// Re-applying a configuration an output already has is not free: compositors
// reject the no-op with "failed to apply configuration".
func TestWaylandControllerLeavesEnabledOutputsAloneOnWake(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	if err := controller.Wake(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertCommands(
		t,
		stub,
		"wlr-randr",
		"wlr-randr --output HDMI-A-2 --on --preferred",
	)
}

func TestWaylandControllerRequiresATool(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	delete(stub.installed, "wlr-randr")

	for name, call := range map[string]func() error{
		"sleep": controller.Sleep,
		"wake":  controller.Wake,
	} {
		if err := call(); !errors.Is(
			err,
			ErrWaylandToolsMissing,
		) {
			t.Fatalf(
				"expected ErrWaylandToolsMissing on %s, got %v",
				name,
				err,
			)
		}
	}

	if len(stub.calls) != 0 {
		t.Fatalf(
			"expected no commands, got %v",
			stub.calls,
		)
	}
}

// A system service inherits no session variables, so the controller has to
// derive both from the user it runs as.
func TestWaylandControllerDerivesTheSessionFromTheServiceUser(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	stub.environment["WAYLAND_ALLOW_OUTPUT_DISABLE"] = "true"

	if err := controller.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEnv(
		t,
		stub,
		"XDG_RUNTIME_DIR=/run/user/1000",
		"WAYLAND_DISPLAY=wayland-1",
	)
}

func TestWaylandControllerKeepsSessionVariablesFromTheEnvironment(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	stub.environment["WAYLAND_ALLOW_OUTPUT_DISABLE"] = "true"

	stub.environment["XDG_RUNTIME_DIR"] = "/run/user/1001"
	stub.environment["WAYLAND_DISPLAY"] = "wayland-0"

	controller.sockets = func(string) ([]string, error) {
		t.Fatal("expected no socket lookup when the environment is set")

		return nil, nil
	}

	if err := controller.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEnv(
		t,
		stub,
		"XDG_RUNTIME_DIR=/run/user/1001",
		"WAYLAND_DISPLAY=wayland-0",
	)
}

func assertEnv(
	t *testing.T,
	stub *waylandCommandStub,
	expected ...string,
) {
	t.Helper()

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

	stub.environment["WAYLAND_ALLOW_OUTPUT_DISABLE"] = "true"

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

func TestWaylandControllerReportsAToolThatVanished(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	stub.environment["WAYLAND_ALLOW_OUTPUT_DISABLE"] = "true"

	stub.errs["wlr-randr"] = &exec.Error{
		Name: "wlr-randr",
		Err:  exec.ErrNotFound,
	}

	err := controller.Sleep()

	if err == nil {
		t.Fatal("expected an error when wlr-randr cannot run")
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

	stub.environment["WAYLAND_ALLOW_OUTPUT_DISABLE"] = "true"

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

// A compositor that refuses to re-enable a disabled output leaves the screen
// dark, so the error has to point at the tool that avoids the problem.
func TestWaylandControllerSuggestsWlopmWhenConfigurationIsRefused(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	stub.outputs["wlr-randr --output HDMI-A-2 --on --preferred"] =
		"failed to apply configuration\n"
	stub.errs["wlr-randr --output HDMI-A-2 --on --preferred"] =
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

// A screen that the fallback cannot switch back on must not be switched off on
// a timer, so the fallback is refused until it is asked for by name.
func TestWaylandControllerRefusesToDisableOutputsByDefault(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	if err := controller.Sleep(); !errors.Is(
		err,
		ErrWlopmRequired,
	) {
		t.Fatalf(
			"expected ErrWlopmRequired, got %v",
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

// Waking is always safe, so it never waits for the opt-in: it is what recovers
// a screen the fallback already switched off.
func TestWaylandControllerWakesWithoutTheOptIn(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	if err := controller.Wake(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertCommands(
		t,
		stub,
		"wlr-randr",
		"wlr-randr --output HDMI-A-2 --on --preferred",
	)
}

// A display in standby stops advertising modes, and then no configuration can
// be applied to its output. The fix is at the screen, so the error has to say
// that rather than relay the compositor's refusal.
func TestWaylandControllerReportsAnOutputWithNoModes(t *testing.T) {
	controller, stub := newStubbedWaylandController()

	stub.outputs["wlr-randr"] = `HDMI-A-1 "Samsung Electric Company SAMSUNG (HDMI-A-1)"
  Enabled: no
`

	err := controller.Wake()

	if !errors.Is(err, ErrOutputNotReporting) {
		t.Fatalf(
			"expected ErrOutputNotReporting, got %v",
			err,
		)
	}

	if !strings.Contains(err.Error(), "HDMI-A-1") {
		t.Fatalf(
			"expected the error to name the output, got %v",
			err,
		)
	}

	assertCommands(t, stub, "wlr-randr")
}
