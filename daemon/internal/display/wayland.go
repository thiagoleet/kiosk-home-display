package display

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrWaylandToolsMissing reports that neither tool the wayland mode can use is
// installed, so nothing on this host can power the screen.
var ErrWaylandToolsMissing = errors.New(
	"neither wlopm nor wlr-randr is installed: install wlopm to power the screen",
)

// WaylandController powers the screen on the compositors that Raspberry Pi OS
// ships since Bookworm (labwc, wayfire), where there is no DPMS and no xset.
//
// It prefers wlopm, which speaks zwlr_output_power_management_v1 — the Wayland
// equivalent of DPMS. The output keeps its mode and position and only the sink
// powers down, so nothing in the session is reconfigured. wlr-randr is the
// fallback for hosts without wlopm, and it works differently: it disables the
// output altogether, which reflows every surface and, on some compositor and
// display combinations, cannot be undone — re-enabling answers "failed to
// apply configuration" and the screen stays dark.
type WaylandController struct {
	command waylandRunner

	// The rest resolves the session and the tooling. They are fields so tests
	// can describe a host without one being present.
	lookPath  func(string) (string, error)
	lookupEnv func(string) (string, bool)
	sockets   func(directory string) ([]string, error)
	uid       int
}

func NewWaylandController() *WaylandController {
	return &WaylandController{
		command:   runCommandWithEnv,
		lookPath:  exec.LookPath,
		lookupEnv: os.LookupEnv,
		sockets:   waylandSockets,
		uid:       os.Getuid(),
	}
}

// Wake re-enables any output that was left disabled before powering the sinks
// back on. Nothing can power-manage a disabled output, so a screen that the
// wlr-randr fallback switched off recovers here instead of needing the session
// restarted.
func (c *WaylandController) Wake() error {
	env, err := c.sessionEnv()
	if err != nil {
		return err
	}

	hasWlopm := c.available("wlopm")
	hasWlrRandr := c.available("wlr-randr")

	if !hasWlopm && !hasWlrRandr {
		return ErrWaylandToolsMissing
	}

	if hasWlrRandr {
		if err := c.enableOutputs(env); err != nil {
			return err
		}
	}

	if !hasWlopm {
		return nil
	}

	return c.setPowerWithWlopm(env, "on")
}

func (c *WaylandController) Sleep() error {
	env, err := c.sessionEnv()
	if err != nil {
		return err
	}

	if c.available("wlopm") {
		return c.setPowerWithWlopm(env, "off")
	}

	if !c.available("wlr-randr") {
		return ErrWaylandToolsMissing
	}

	log.Println(
		"[DISPLAY] wlopm is not installed, disabling the output with wlr-randr instead; some compositors refuse to re-enable it",
	)

	return c.disableOutputs(env)
}

// available reports whether a tool is on PATH. Looked up per call, so
// installing wlopm takes effect without restarting the service.
func (c *WaylandController) available(tool string) bool {
	_, err := c.lookPath(tool)

	return err == nil
}

// SetBrightness is not available: neither protocol exposes a brightness or
// gamma channel. The error is the sentinel the daemon tolerates at startup, so
// a wayland kiosk boots with the brightness setting simply ignored.
func (c *WaylandController) SetBrightness(level int) error {
	return fmt.Errorf(
		"%w: the wayland display mode has no brightness control",
		ErrBrightnessUnsupported,
	)
}

func (c *WaylandController) setPowerWithWlopm(
	env []string,
	state string,
) error {
	// wlopm takes "*" as every output. The argument reaches it untouched: no
	// shell is involved, so there is nothing to glob it against.
	result, err := c.command(
		env,
		"wlopm",
		"--"+state,
		"*",
	)

	if err != nil {
		return fmt.Errorf(
			"wlopm --%s: %w",
			state,
			waylandToolError("wlopm", env, result, err),
		)
	}

	return nil
}

func (c *WaylandController) disableOutputs(
	env []string,
) error {
	outputs, err := c.outputs(env)
	if err != nil {
		return err
	}

	for _, output := range outputs {
		if !output.enabled {
			continue
		}

		if err := c.applyOutput(
			env,
			output.name,
			"--off",
		); err != nil {
			return err
		}
	}

	return nil
}

// enableOutputs re-enables the outputs that report "Enabled: no" and leaves the
// rest alone. Re-applying a configuration an output already has is not free:
// compositors have been seen to reject the no-op with "failed to apply
// configuration".
func (c *WaylandController) enableOutputs(
	env []string,
) error {
	outputs, err := c.outputs(env)
	if err != nil {
		return err
	}

	for _, output := range outputs {
		if output.enabled {
			continue
		}

		if err := c.applyOutput(
			env,
			output.name,
			"--on",
		); err != nil {
			return err
		}
	}

	return nil
}

func (c *WaylandController) applyOutput(
	env []string,
	output string,
	flag string,
) error {
	result, err := c.command(
		env,
		"wlr-randr",
		"--output",
		output,
		flag,
	)

	if err != nil {
		return fmt.Errorf(
			"apply %s to output %s: %w",
			flag,
			output,
			waylandToolError("wlr-randr", env, result, err),
		)
	}

	return nil
}

// sessionEnv supplies the two variables the tools need. A system service starts
// outside the graphical session and inherits neither, but it does run as the
// user that owns that session: the runtime directory is /run/user/<uid> and the
// compositor's socket sits inside it. Values already in the environment win, so
// the env file can still pin them.
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

// waylandOutput is one entry of the wlr-randr listing. A disabled output stays
// in the listing, reporting "Enabled: no", which is what makes it possible to
// tell a screen that is merely asleep from one whose output was switched off.
type waylandOutput struct {
	name    string
	enabled bool
}

func (c *WaylandController) outputs(
	env []string,
) ([]waylandOutput, error) {
	result, err := c.command(env, "wlr-randr")
	if err != nil {
		return nil, waylandToolError(
			"wlr-randr",
			env,
			result,
			err,
		)
	}

	var outputs []waylandOutput

	for _, line := range strings.Split(result, "\n") {
		if line == "" {
			continue
		}

		// An output name opens a block in the first column; every property
		// wlr-randr prints underneath it is indented.
		if line[0] != ' ' && line[0] != '\t' {
			fields := strings.Fields(line)

			if len(fields) == 0 {
				continue
			}

			outputs = append(outputs, waylandOutput{
				name: fields[0],
			})

			continue
		}

		if len(outputs) == 0 {
			continue
		}

		if enabled, ok := parseEnabled(line); ok {
			outputs[len(outputs)-1].enabled = enabled
		}
	}

	if len(outputs) == 0 {
		return nil, errors.New(
			"wlr-randr reported no outputs",
		)
	}

	return outputs, nil
}

func parseEnabled(line string) (bool, bool) {
	value, found := strings.CutPrefix(
		strings.TrimSpace(line),
		"Enabled:",
	)

	if !found {
		return false, false
	}

	return strings.TrimSpace(value) == "yes", true
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

// waylandToolError names the ways these tools fail on a kiosk: the package is
// missing, the socket they were pointed at is not the compositor's, or the
// compositor turns the request down. The resolved socket goes into the message,
// since the daemon picks it without being told.
func waylandToolError(
	tool string,
	env []string,
	output string,
	err error,
) error {
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf(
			"%s is not installed, install the %s package: %w",
			tool,
			tool,
			err,
		)
	}

	if strings.Contains(output, "failed to connect") {
		return fmt.Errorf(
			"%s could not connect to %s in %s, check that the compositor runs as the service user: %w",
			tool,
			envValue(env, "WAYLAND_DISPLAY"),
			envValue(env, "XDG_RUNTIME_DIR"),
			err,
		)
	}

	if strings.Contains(output, "doesn't support") ||
		strings.Contains(output, "not supported") {
		return fmt.Errorf(
			"the compositor does not support the protocol %s needs: %w",
			tool,
			err,
		)
	}

	// Disabling an output is not always reversible: the compositor can refuse
	// to re-enable it, and then the screen stays dark. wlopm never gets here,
	// because powering a sink down leaves the output configured.
	if strings.Contains(output, "failed to apply configuration") {
		return fmt.Errorf(
			"the compositor refused the output configuration; install wlopm, which powers the screen without reconfiguring outputs: %w",
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
