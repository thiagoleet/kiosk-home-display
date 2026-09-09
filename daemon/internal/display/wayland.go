package display

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WaylandController drives the screen with wlr-randr, the wlr-output-management
// client that the compositors on Raspberry Pi OS Bookworm and later understand
// (labwc, wayfire). Wayland has no DPMS and no xset, so sleeping disables the
// output and waking re-enables it, which powers the HDMI link down all the
// same.
type WaylandController struct {
	command waylandRunner

	// The rest resolves the session wlr-randr has to talk to. They are fields
	// so tests can describe a session without one being present.
	lookupEnv func(string) (string, bool)
	sockets   func(directory string) ([]string, error)
	uid       int
}

func NewWaylandController() *WaylandController {
	return &WaylandController{
		command:   runCommandWithEnv,
		lookupEnv: os.LookupEnv,
		sockets:   waylandSockets,
		uid:       os.Getuid(),
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
	env, err := c.sessionEnv()
	if err != nil {
		return err
	}

	outputs, err := c.outputs(env)
	if err != nil {
		return err
	}

	for _, output := range outputs {
		result, err := c.command(
			env,
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
				wlrRandrError(env, result, err),
			)
		}
	}

	return nil
}

// sessionEnv supplies the two variables wlr-randr needs. A system service
// starts outside the graphical session and inherits neither, but it does run as
// the user that owns that session: the runtime directory is /run/user/<uid> and
// the compositor's socket sits inside it. Values already in the environment win,
// so the env file can still pin them.
func (c *WaylandController) sessionEnv() ([]string, error) {
	runtimeDir, ok := c.lookupEnv("XDG_RUNTIME_DIR")

	if !ok || runtimeDir == "" {
		runtimeDir = fmt.Sprintf("/run/user/%d", c.uid)
	}

	env := []string{
		"XDG_RUNTIME_DIR=" + runtimeDir,
	}

	if display, ok := c.lookupEnv(
		"WAYLAND_DISPLAY",
	); ok && display != "" {
		return append(
			env,
			"WAYLAND_DISPLAY="+display,
		), nil
	}

	sockets, err := c.sockets(runtimeDir)
	if err != nil {
		return nil, fmt.Errorf(
			"look for a wayland socket in %s: %w",
			runtimeDir,
			err,
		)
	}

	if len(sockets) == 0 {
		return nil, fmt.Errorf(
			"no wayland socket in %s: the compositor is not running as this user, or the session is not wayland",
			runtimeDir,
		)
	}

	return append(
		env,
		"WAYLAND_DISPLAY="+filepath.Base(sockets[0]),
	), nil
}

// outputs lists every output wlr-randr knows about, enabled or not. Waking has
// to reach the outputs that sleeping disabled, and a disabled output still
// appears in the listing, reporting "Enabled: no".
func (c *WaylandController) outputs(
	env []string,
) ([]string, error) {
	result, err := c.command(env, "wlr-randr")
	if err != nil {
		return nil, wlrRandrError(env, result, err)
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

// waylandSockets lists the compositor sockets in a runtime directory. Each
// socket comes with a "wayland-0.lock" companion that is not a socket, so the
// lock files are dropped.
func waylandSockets(
	directory string,
) ([]string, error) {
	matches, err := filepath.Glob(
		filepath.Join(directory, "wayland-*"),
	)

	if err != nil {
		return nil, err
	}

	var sockets []string

	for _, match := range matches {
		if strings.HasSuffix(match, ".lock") {
			continue
		}

		sockets = append(sockets, match)
	}

	return sockets, nil
}

// wlrRandrError names the ways wlr-randr fails on a kiosk: the package is
// missing, the socket it was pointed at is not the compositor's, or the
// compositor refuses to manage outputs. The resolved socket goes into the
// message, since the daemon picks it without being told.
func wlrRandrError(
	env []string,
	output string,
	err error,
) error {
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

	if strings.Contains(output, "failed to connect") {
		return fmt.Errorf(
			"wlr-randr could not connect to %s in %s, check that the compositor runs as the service user: %w",
			envValue(env, "WAYLAND_DISPLAY"),
			envValue(env, "XDG_RUNTIME_DIR"),
			err,
		)
	}

	return err
}

func envValue(env []string, name string) string {
	prefix := name + "="

	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return strings.TrimPrefix(entry, prefix)
		}
	}

	return "unset " + name
}
