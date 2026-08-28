package http

import (
	"encoding/json"
	"net/http"

	"github.com/thiagoleet/kiosk-home-display/internal/events"
	"github.com/thiagoleet/kiosk-home-display/internal/i18n"
)

type NotificationHandler struct {
	bus   *events.Bus
	texts i18n.Catalog
}

type notificationRequest struct {
	Context events.NotificationContext `json:"context"`
	Title   string                     `json:"title"`
	Message string                     `json:"message"`
	Level   events.NotificationLevel   `json:"level"`
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

func (h *NotificationHandler) Notify(
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

	var request notificationRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&request); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)

		return
	}

	events.PublishNotification(
		h.bus,
		events.NewNotification(
			request.Context,
			request.Title,
			request.Message,
			request.Level,
			5000,
		),
	)

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})

}
