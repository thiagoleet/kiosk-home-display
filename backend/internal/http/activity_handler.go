package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/thiagoleet/kiosk-home-display/internal/activity"
)

type ActivityHandler struct {
	repository activity.Repository
}

func NewActivityHandler(
	repository activity.Repository,
) *ActivityHandler {
	return &ActivityHandler{
		repository: repository,
	}
}

func (h *ActivityHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	limit := 5

	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed <= 0 {
			http.Error(
				w,
				"invalid limit",
				http.StatusBadRequest,
			)

			return
		}

		limit = parsed
	}

	activities, err := h.repository.List(
		r.Context(),
		limit,
	)
	if err != nil {
		http.Error(
			w,
			"failed to list activities",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(
		activities,
	); err != nil {
		return
	}
}
