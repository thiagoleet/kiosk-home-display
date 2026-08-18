package display

import "log"

type VirtualController struct {
	state      State
	brightness int
}

func NewVirtualController() *VirtualController {
	return &VirtualController{
		state:      StateOn,
		brightness: 100,
	}
}

func (c *VirtualController) Wake() error {
	c.state = StateOn

	log.Println("[DISPLAY] wake")

	return nil
}

func (c *VirtualController) Sleep() error {
	c.state = StateOff

	log.Println("[DISPLAY] sleep")

	return nil
}

func (c *VirtualController) SetBrightness(level int) error {
	c.brightness = level

	log.Printf("[DISPLAY] set brightness to %d%%", level)

	return nil
}
