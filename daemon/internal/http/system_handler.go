package http

import (
	nethttp "net/http"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type SystemHandler struct {
	bus *events.Bus
}

func NewSystemHandler(
	bus *events.Bus,
) *SystemHandler {
	return &SystemHandler{
		bus: bus,
	}
}

func (h *SystemHandler) Refresh(
	w nethttp.ResponseWriter,
	r *nethttp.Request,
) {
	if r.Method != nethttp.MethodPost {
		nethttp.Error(
			w,
			"method not allowed",
			nethttp.StatusMethodNotAllowed,
		)

		return
	}

	h.bus.Publish(events.Event{
		Type: events.EventSystemRefresh,
	})

	writeJSON(w, nethttp.StatusOK, map[string]any{
		"status": "ok",
	})
}
