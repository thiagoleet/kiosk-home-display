package display

import (
	"fmt"
	"os"
	"os/exec"
)

// commandRunner runs an external command and returns its combined output.
// The field exists so tests can drive a controller without a display session.
type commandRunner func(
	name string,
	args ...string,
) (string, error)

// waylandRunner runs a command with extra environment variables. wlr-randr
// needs the session variables that a system service never inherits, so the
// wayland controller supplies them per call.
type waylandRunner func(
	env []string,
	name string,
	args ...string,
) (string, error)

func runCommand(
	name string,
	args ...string,
) (string, error) {
	return runCommandWithEnv(nil, name, args...)
}

func runCommandWithEnv(
	env []string,
	name string,
	args ...string,
) (string, error) {
	cmd := exec.Command(name, args...)

	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

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
