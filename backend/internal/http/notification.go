package http

import (
	"net/http"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
)

type NotificationHandler struct {
	bus *events.Bus
}

func NewNotificationHandler(
	bus *events.Bus,
) *NotificationHandler {
	return &NotificationHandler{
		bus: bus,
	}
}

func (h *NotificationHandler) Test(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}

	events.PublishNotification(
		h.bus,
		events.NewNotification(
			"Test notification",
			"This notification came from the Go backend.",
			events.NotificationInfo,
			5000,
		),
	)

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
