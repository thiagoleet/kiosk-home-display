package display

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

const xrandrQueryOutput = `Screen 0: minimum 320 x 200, current 1920 x 1080, maximum 16384 x 16384
HDMI-1 connected primary 1920x1080+0+0 (normal left inverted right x axis y axis) 698mm x 392mm
   1920x1080     60.00*+  50.00    59.94
HDMI-2 disconnected (normal left inverted right x axis y axis)
`

type recordedCommand struct {
	name string
	args []string
}

type commandStub struct {
	calls   []recordedCommand
	outputs map[string]string
	errs    map[string]error
}

func newCommandStub() *commandStub {
	return &commandStub{
		outputs: map[string]string{
			"xrandr --query": xrandrQueryOutput,
		},
		errs: map[string]error{},
	}
}

func (s *commandStub) run(
	name string,
	args ...string,
) (string, error) {
	s.calls = append(s.calls, recordedCommand{
		name: name,
		args: args,
	})

	key := strings.Join(
		append([]string{name}, args...),
		" ",
	)

	return s.outputs[key], s.errs[key]
}

func newStubbedLinuxController() (
	*LinuxController,
	*commandStub,
) {
	stub := newCommandStub()

	controller := NewLinuxController()
	controller.command = stub.run

	return controller, stub
}

func TestLinuxControllerSetsBrightnessOnConnectedOutputs(t *testing.T) {
	controller, stub := newStubbedLinuxController()

	if err := controller.SetBrightness(40); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stub.calls) != 2 {
		t.Fatalf(
			"expected 2 commands, got %d: %v",
			len(stub.calls),
			stub.calls,
		)
	}

	applied := stub.calls[1]

	if applied.name != "xrandr" {
		t.Fatalf(
			"expected xrandr, got %q",
			applied.name,
		)
	}

	want := "--output HDMI-1 --brightness 0.40"

	if got := strings.Join(applied.args, " "); got != want {
		t.Fatalf(
			"expected args %q, got %q",
			want,
			got,
		)
	}
}

func TestLinuxControllerRejectsBrightnessOutOfRange(t *testing.T) {
	controller, stub := newStubbedLinuxController()

	if err := controller.SetBrightness(101); err == nil {
		t.Fatal("expected an error for brightness above 100")
	}

	if len(stub.calls) != 0 {
		t.Fatalf(
			"expected no commands, got %v",
			stub.calls,
		)
	}
}

func TestLinuxControllerReportsMissingXrandrAsUnsupported(t *testing.T) {
	controller, stub := newStubbedLinuxController()

	stub.errs["xrandr --query"] = &exec.Error{
		Name: "xrandr",
		Err:  exec.ErrNotFound,
	}

	err := controller.SetBrightness(50)

	if !errors.Is(err, ErrBrightnessUnsupported) {
		t.Fatalf(
			"expected ErrBrightnessUnsupported, got %v",
			err,
		)
	}
}

func TestLinuxControllerReportsNoConnectedOutputAsUnsupported(t *testing.T) {
	controller, stub := newStubbedLinuxController()

	stub.outputs["xrandr --query"] =
		"HDMI-1 disconnected (normal left inverted right)\n"

	err := controller.SetBrightness(50)

	if !errors.Is(err, ErrBrightnessUnsupported) {
		t.Fatalf(
			"expected ErrBrightnessUnsupported, got %v",
			err,
		)
	}
}

func TestLinuxControllerSkipsConnectedOutputWithoutMode(t *testing.T) {
	controller, stub := newStubbedLinuxController()

	stub.outputs["xrandr --query"] = `Screen 0: minimum 320 x 200, current 1920 x 1080, maximum 16384 x 16384
HDMI-1 connected primary 1920x1080+0+0 (normal left inverted right x axis y axis) 698mm x 392mm
HDMI-2 connected (normal left inverted right x axis y axis) 530mm x 300mm
`

	if err := controller.SetBrightness(40); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stub.calls) != 2 {
		t.Fatalf(
			"expected only the active output to be adjusted, got %v",
			stub.calls,
		)
	}

	want := "--output HDMI-1 --brightness 0.40"

	if got := strings.Join(stub.calls[1].args, " "); got != want {
		t.Fatalf(
			"expected args %q, got %q",
			want,
			got,
		)
	}
}

func TestLinuxControllerReportsInactiveOutputsAsUnsupported(t *testing.T) {
	controller, stub := newStubbedLinuxController()

	stub.outputs["xrandr --query"] =
		"HDMI-2 connected (normal left inverted right) 530mm x 300mm\n"

	err := controller.SetBrightness(50)

	if !errors.Is(err, ErrBrightnessUnsupported) {
		t.Fatalf(
			"expected ErrBrightnessUnsupported, got %v",
			err,
		)
	}
}

func TestLinuxControllerWakesAndSleepsWithXset(t *testing.T) {
	controller, stub := newStubbedLinuxController()

	if err := controller.Wake(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := controller.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"xset dpms force on",
		"xset q",
		"xset dpms force off",
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

func TestLinuxControllerEnablesDPMSBeforeSleeping(t *testing.T) {
	controller, stub := newStubbedLinuxController()

	stub.outputs["xset q"] = `DPMS (Energy Star):
  Standby: 600    Suspend: 600    Off: 600
  DPMS is Disabled
  Monitor is On
`

	if err := controller.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"xset q",
		"xset +dpms",
		"xset dpms 0 0 0",
		"xset dpms force off",
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

func TestLinuxControllerKeepsDPMSSettingsWhenEnabled(t *testing.T) {
	controller, stub := newStubbedLinuxController()

	stub.outputs["xset q"] = `DPMS (Energy Star):
  Standby: 0    Suspend: 0    Off: 0
  DPMS is Enabled
  Monitor is On
`

	if err := controller.Sleep(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, call := range stub.calls {
		if len(call.args) == 1 && call.args[0] == "+dpms" {
			t.Fatalf(
				"expected no dpms change, got %v",
				stub.calls,
			)
		}
	}
}

func TestLinuxControllerReportsMissingXset(t *testing.T) {
	controller, stub := newStubbedLinuxController()

	stub.errs["xset q"] = &exec.Error{
		Name: "xset",
		Err:  exec.ErrNotFound,
	}

	err := controller.Sleep()

	if err == nil {
		t.Fatal("expected an error when xset is missing")
	}

	if !strings.Contains(err.Error(), "x11-xserver-utils") {
		t.Fatalf(
			"expected the error to name the package, got %v",
			err,
		)
	}
}

func TestLinuxControllerReportsUnreachableXSession(t *testing.T) {
	controller, stub := newStubbedLinuxController()

	stub.outputs["xset q"] = "xset:  unable to open display \"\"\n"
	stub.errs["xset q"] = errors.New("exit status 1")

	err := controller.Sleep()

	if err == nil {
		t.Fatal("expected an error when the X session is unreachable")
	}

	if !strings.Contains(err.Error(), "XAUTHORITY") {
		t.Fatalf(
			"expected the error to name the missing environment, got %v",
			err,
		)
	}
}
