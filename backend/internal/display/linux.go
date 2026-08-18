package display

import (
	"fmt"
	"os/exec"
)

type LinuxController struct{}

func NewLinuxController() *LinuxController {
	return &LinuxController{}
}

func (c *LinuxController) Wake() error {
	return runCommand("xset", "dpms", "force", "on")
}

func (c *LinuxController) Sleep() error {
	return runCommand("xset", "dpms", "force", "off")
}

// func (c *LinuxController) Dim() error {
// 	return fmt.Errorf("display dimming is not implemented")
// }

func (c *LinuxController) SetBrightness(level int) error {
	return fmt.Errorf("display brightness adjustment is not implemented")
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf(
			"command %s failed: %w: %s",
			name,
			err,
			string(output),
		)
	}

	return nil

}
