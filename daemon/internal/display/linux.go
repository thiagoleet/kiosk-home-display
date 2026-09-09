package display

import (
	"errors"
	"fmt"
	"log"
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

type LinuxController struct {
	command commandRunner
}

func NewLinuxController() *LinuxController {
	return &LinuxController{
		command: runCommand,
	}
}

func (c *LinuxController) Wake() error {
	output, err := c.command(
		"xset",
		"dpms",
		"force",
		"on",
	)

	if err != nil {
		return xsetError(output, err)
	}

	return nil
}

// Sleep powers the screen down through DPMS. It enables DPMS first, because a
// kiosk session usually turns it off to stop the screen blanking on its own —
// raspi-config writes "X -s 0 -dpms" for exactly that. With DPMS disabled,
// "xset dpms force off" is a silent no-op: it exits 0 and the screen stays lit.
func (c *LinuxController) Sleep() error {
	if err := c.enableDPMS(); err != nil {
		return err
	}

	output, err := c.command(
		"xset",
		"dpms",
		"force",
		"off",
	)

	if err != nil {
		return xsetError(output, err)
	}

	return nil
}

// enableDPMS turns the DPMS extension on when the X server reports it off, and
// zeroes its three timeouts. The timeouts are what the kiosk session wanted
// gone: with them at zero the screen only ever powers down when this daemon
// asks it to, so enabling DPMS does not bring blanking back.
func (c *LinuxController) enableDPMS() error {
	output, err := c.command("xset", "q")
	if err != nil {
		return xsetError(output, err)
	}

	if !strings.Contains(output, "DPMS is Disabled") {
		return nil
	}

	log.Println(
		"[DISPLAY] DPMS is disabled, enabling it so the screen can power down",
	)

	for _, args := range [][]string{
		{"+dpms"},
		{"dpms", "0", "0", "0"},
	} {
		output, err := c.command("xset", args...)
		if err != nil {
			return fmt.Errorf(
				"enable dpms: %w",
				xsetError(output, err),
			)
		}
	}

	return nil
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

// xsetError names the two ways xset fails on a kiosk: the package is missing,
// or the daemon has no X session to talk to. Both look like a plain exit status
// in the journal otherwise, which is the hardest part of the problem to spot.
func xsetError(output string, err error) error {
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf(
			"xset is not installed, install x11-xserver-utils: %w",
			err,
		)
	}

	if strings.Contains(output, "unable to open display") {
		return fmt.Errorf(
			"xset cannot reach the X session, check DISPLAY and XAUTHORITY: %w",
			err,
		)
	}

	return err
}
