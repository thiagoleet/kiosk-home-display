package http

import (
	"errors"
	nethttp "net/http"
	"strconv"

	"github.com/thiagoleet/kiosk-home-display/internal/news"
)

type NewsHandler struct {
	service *news.Service
}

func NewNewsHandler(
	service *news.Service,
) *NewsHandler {
	return &NewsHandler{
		service: service,
	}
}

// List serves the persisted articles, newest first. It never fetches: the
// scheduler keeps the database current, so a slow feed cannot stall the
// slideshow's request.
func (h *NewsHandler) List(
	w nethttp.ResponseWriter,
	r *nethttp.Request,
) {
	if r.Method != nethttp.MethodGet {
		nethttp.Error(
			w,
			"method not allowed",
			nethttp.StatusMethodNotAllowed,
		)

		return
	}

	limit := 0

	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed <= 0 {
			nethttp.Error(
				w,
				"invalid limit",
				nethttp.StatusBadRequest,
			)

			return
		}

		limit = parsed
	}

	feed, err := h.service.List(r.Context(), limit)
	if err != nil {
		if errors.Is(err, news.ErrDisabled) {
			nethttp.Error(
				w,
				"News feed is not enabled for this device",
				nethttp.StatusServiceUnavailable,
			)

			return
		}

		nethttp.Error(
			w,
			"failed to list news",
			nethttp.StatusInternalServerError,
		)

		return
	}

	writeJSON(
		w,
		nethttp.StatusOK,
		feed,
	)
}
