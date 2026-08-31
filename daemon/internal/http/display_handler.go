package http

import (
	"encoding/json"
	nethttp "net/http"
	"strconv"

	"github.com/thiagoleet/kiosk-home-display/internal/display"
)

type DisplayHandler struct {
	display *display.Manager
}

func NewDisplayHandler(
	displayManager *display.Manager,
) *DisplayHandler {
	return &DisplayHandler{
		display: displayManager,
	}
}

func (h *DisplayHandler) Sleep(
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

	if err := h.display.Sleep(); err != nil {
		nethttp.Error(
			w,
			err.Error(),
			nethttp.StatusInternalServerError,
		)

		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (h *DisplayHandler) Wake(
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

	if err := h.display.Wake(); err != nil {
		nethttp.Error(
			w,
			err.Error(),
			nethttp.StatusInternalServerError,
		)

		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (h *DisplayHandler) Brightness(
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

	level, err := strconv.Atoi(
		r.URL.Query().Get("level"),
	)

	if err != nil {
		nethttp.Error(
			w,
			"invalid brightness level",
			nethttp.StatusBadRequest,
		)

		return
	}

	if err := h.display.SetBrightness(level); err != nil {
		nethttp.Error(
			w,
			err.Error(),
			nethttp.StatusBadRequest,
		)

		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]any{
		"status": "ok",
	})
}

func writeJSON(
	w nethttp.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}
