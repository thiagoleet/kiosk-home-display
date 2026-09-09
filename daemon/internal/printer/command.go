package printer

import (
	"fmt"
	"os/exec"
)

// commandRunner runs an external command and returns its combined output.
// The field exists so tests can drive a source without a CUPS installation.
type commandRunner func(
	name string,
	args ...string,
) (string, error)

func runCommand(
	name string,
	args ...string,
) (string, error) {
	output, err := exec.Command(
		name,
		args...,
	).CombinedOutput()

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
