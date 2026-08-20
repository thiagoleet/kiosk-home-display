package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/thiagoleet/kiosk-home-display/internal/weather"
)

type WeatherHandler struct {
	service *weather.Service
}

func NewWeatherHandler(
	service *weather.Service,
) *WeatherHandler {
	return &WeatherHandler{
		service: service,
	}
}

func (h *WeatherHandler) Current(
	w nethttp.ResponseWriter,
	r *nethttp.Request,
) {
	currentWeather, err := h.service.GetCurrent(
		r.Context(),
	)
	if err != nil {
		if errors.Is(err, weather.ErrDisabled) {
			nethttp.Error(
				w,
				"Weather Forecast is not enabled for this device",
				nethttp.StatusServiceUnavailable,
			)

			return
		}

		nethttp.Error(
			w,
			"failed to get weather",
			nethttp.StatusBadGateway,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(nethttp.StatusOK)

	if err := json.NewEncoder(w).Encode(
		currentWeather,
	); err != nil {
		return
	}
}
