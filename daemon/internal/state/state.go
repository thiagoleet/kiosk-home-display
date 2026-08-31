package state

import "github.com/thiagoleet/kiosk-home-display/internal/display"

type State struct {
	Display DisplayState `json:"display"`
}

type DisplayState struct {
	Power      display.State `json:"power"`
	Brightness int           `json:"brightness"`
}
