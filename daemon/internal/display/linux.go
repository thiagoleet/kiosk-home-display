package display

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// activeGeometry matches the "1920x1080+0+0" field that xrandr prints for an
// output that currently drives a CRTC. A connected output without one is not
// configurable: xrandr answers --brightness with a BadMatch on RRSetCrtcConfig.
var activeGeometry = regexp.MustCompile(
	`^[0-9]+x[0-9]+\+[0-9]+\+[0-9]+$`,
)

// commandRunner runs an external command and returns its combined output.
// The field exists so tests can drive the controller without an X session.
type commandRunner func(
	name string,
	args ...string,
) (string, error)

type LinuxController struct {
	command commandRunner
}

func NewLinuxController() *LinuxController {
	return &LinuxController{
		command: runCommand,
	}
}

func (c *LinuxController) Wake() error {
	_, err := c.command(
		"xset",
		"dpms",
		"force",
		"on",
	)

	return err
}

func (c *LinuxController) Sleep() error {
	_, err := c.command(
		"xset",
		"dpms",
		"force",
		"off",
	)

	return err
}

// SetBrightness dims every connected output through xrandr. It is a software
// gamma adjustment, not a backlight change, so it works on HDMI screens that
// expose no backlight device. Level 0 renders the screen black without
// powering it off; use Sleep for that.
func (c *LinuxController) SetBrightness(level int) error {
	if level < 0 || level > 100 {
		return fmt.Errorf(
			"brightness must be between 0 and 100, got %d",
			level,
		)
	}

	outputs, err := c.connectedOutputs()
	if err != nil {
		return err
	}

	value := strconv.FormatFloat(
		float64(level)/100,
		'f',
		2,
		64,
	)

	for _, output := range outputs {
		if _, err := c.command(
			"xrandr",
			"--output",
			output,
			"--brightness",
			value,
		); err != nil {
			return fmt.Errorf(
				"set brightness on output %s: %w",
				output,
				err,
			)
		}
	}

	return nil
}

// connectedOutputs lists the xrandr outputs that have a screen attached and a
// mode applied. Outputs that are merely connected are skipped: swapping a
// display leaves the new output enumerated but unconfigured until a mode is
// set, and asking xrandr to adjust its brightness fails with a BadMatch.
func (c *LinuxController) connectedOutputs() ([]string, error) {
	output, err := c.command("xrandr", "--query")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf(
				"%w: xrandr is not installed",
				ErrBrightnessUnsupported,
			)
		}

		return nil, fmt.Errorf(
			"query xrandr outputs: %w",
			err,
		)
	}

	var (
		outputs   []string
		connected int
	)

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		if fields[1] != "connected" {
			continue
		}

		connected++

		if !hasActiveMode(fields[2:]) {
			continue
		}

		outputs = append(outputs, fields[0])
	}

	if len(outputs) == 0 {
		if connected > 0 {
			return nil, fmt.Errorf(
				"%w: no connected output has a mode set",
				ErrBrightnessUnsupported,
			)
		}

		return nil, fmt.Errorf(
			"%w: no connected output found",
			ErrBrightnessUnsupported,
		)
	}

	return outputs, nil
}

// hasActiveMode reports whether the fields that follow "connected" carry the
// geometry xrandr prints for an output that is driving a CRTC. The geometry
// sits before the "(normal left inverted ...)" property list, optionally
// preceded by "primary".
func hasActiveMode(fields []string) bool {
	for _, field := range fields {
		if strings.HasPrefix(field, "(") {
			return false
		}

		if activeGeometry.MatchString(field) {
			return true
		}
	}

	return false
}

func runCommand(
	name string,
	args ...string,
) (string, error) {
	cmd := exec.Command(name, args...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return string(output), fmt.Errorf(
			"command %s failed: %w: %s",
			name,
			err,
			string(output),
		)
	}

	return string(output), nil
}
