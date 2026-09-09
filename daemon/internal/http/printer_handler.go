package http

import (
	"encoding/json"
	"errors"
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
			printErrorStatus(err),
		)

		return
	}

	writeJSON(
		w,
		nethttp.StatusAccepted,
		job,
	)
}

// printErrorStatus separates a caller that asked for something impossible from
// a printer that cannot take the job right now.
func printErrorStatus(err error) int {
	switch {
	case errors.Is(err, printer.ErrInvalidJobName):
		return nethttp.StatusBadRequest

	case errors.Is(err, printer.ErrMonitored):
		return nethttp.StatusConflict

	case errors.Is(err, printer.ErrBusy):
		return nethttp.StatusConflict

	default:
		return nethttp.StatusInternalServerError
	}
}
