package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/thiagoleet/kiosk-home-display/internal/printer"
)

type PrinterHandler struct {
	printer *printer.Manager
}

func NewPrinterHandler(
	printerManager *printer.Manager,
) *PrinterHandler {
	return &PrinterHandler{
		printer: printerManager,
	}
}

type printRequest struct {
	Name string `json:"name"`
}

func (h *PrinterHandler) Print(
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

	var request printRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&request); err != nil {
		nethttp.Error(
			w,
			"invalid request body",
			nethttp.StatusBadRequest,
		)

		return
	}

	job, err := h.printer.Print(
		request.Name,
	)

	if err != nil {
		nethttp.Error(
			w,
			err.Error(),
			nethttp.StatusConflict,
		)

		return
	}

	writeJSON(
		w,
		nethttp.StatusAccepted,
		job,
	)
}
