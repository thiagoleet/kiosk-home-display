package http

import (
	"net/http"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
	"github.com/thiagoleet/kiosk-home-display/internal/i18n"
)

type NotificationHandler struct {
	bus   *events.Bus
	texts i18n.Catalog
}

func NewNotificationHandler(
	bus *events.Bus,
	texts i18n.Catalog,
) *NotificationHandler {
	return &NotificationHandler{
		bus:   bus,
		texts: texts,
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
			events.NotificationContextSystem,
			h.texts.Text(i18n.KeyTestNotificationTitle),
			h.texts.Text(i18n.KeyTestNotificationMessage),
			events.NotificationInfo,
			5000,
		),
	)

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
