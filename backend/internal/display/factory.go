package display

import "fmt"

func NewController(mode string) (Controller, error) {
	switch mode {
	case "virtual":
		return NewVirtualController(), nil

	case "linux":
		return NewLinuxController(), nil

	default:
		return nil, fmt.Errorf("unknown display mode: %s", mode)

	}
}
