package display

import "errors"

// ErrBrightnessUnsupported is returned by controllers that drive hardware
// without a brightness channel. Callers that only want a best-effort
// adjustment can ignore it with errors.Is.
var ErrBrightnessUnsupported = errors.New(
	"display brightness adjustment is not supported",
)
