package display

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// WaylandController drives the screen with wlr-randr, the wlr-output-management
// client that the compositors on Raspberry Pi OS Bookworm and later understand
// (labwc, wayfire). Wayland has no DPMS and no xset, so sleeping disables the
// output and waking re-enables it, which powers the HDMI link down all the
// same.
type WaylandController struct {
	command commandRunner
}

func NewWaylandController() *WaylandController {
	return &WaylandController{
		command: runCommand,
	}
}

func (c *WaylandController) Wake() error {
	return c.applyToOutputs("--on")
}

func (c *WaylandController) Sleep() error {
	return c.applyToOutputs("--off")
}

// SetBrightness is not available: wlr-output-management exposes no brightness
// or gamma channel. The error is the sentinel the daemon tolerates at startup,
// so a wayland kiosk boots with the brightness setting simply ignored.
func (c *WaylandController) SetBrightness(level int) error {
	return fmt.Errorf(
		"%w: the wayland display mode has no brightness control",
		ErrBrightnessUnsupported,
	)
}

func (c *WaylandController) applyToOutputs(state string) error {
	outputs, err := c.outputs()
	if err != nil {
		return err
	}

	for _, output := range outputs {
		result, err := c.command(
			"wlr-randr",
			"--output",
			output,
			state,
		)

		if err != nil {
			return fmt.Errorf(
				"apply %s to output %s: %w",
				state,
				output,
				wlrRandrError(result, err),
			)
		}
	}

	return nil
}

// outputs lists every output wlr-randr knows about, enabled or not. Waking has
// to reach the outputs that sleeping disabled, and a disabled output still
// appears in the listing, reporting "Enabled: no".
func (c *WaylandController) outputs() ([]string, error) {
	result, err := c.command("wlr-randr")
	if err != nil {
		return nil, wlrRandrError(result, err)
	}

	var outputs []string

	for _, line := range strings.Split(result, "\n") {
		// An output name opens a block in the first column; every property
		// wlr-randr prints underneath it is indented.
		if line == "" || line[0] == ' ' || line[0] == '\t' {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) == 0 {
			continue
		}

		outputs = append(outputs, fields[0])
	}

	if len(outputs) == 0 {
		return nil, errors.New(
			"wlr-randr reported no outputs",
		)
	}

	return outputs, nil
}

// wlrRandrError names the two ways wlr-randr fails on a kiosk: the package is
// missing, or the daemon runs outside the compositor session and never reaches
// the Wayland socket.
func wlrRandrError(output string, err error) error {
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf(
			"wlr-randr is not installed, install the wlr-randr package: %w",
			err,
		)
	}

	if strings.Contains(output, "compositor doesn't support") {
		return fmt.Errorf(
			"the compositor does not support wlr-output-management: %w",
			err,
		)
	}

	if strings.Contains(output, "failed to connect") ||
		strings.Contains(output, "WAYLAND_DISPLAY") {
		return fmt.Errorf(
			"wlr-randr cannot reach the compositor, check WAYLAND_DISPLAY and XDG_RUNTIME_DIR: %w",
			err,
		)
	}

	return err
}
