package http

import (
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
		writeWeatherError(w, err)

		return
	}

	writeJSON(
		w,
		nethttp.StatusOK,
		currentWeather,
	)
}

// Forecast serves the daily outlook, today first.
func (h *WeatherHandler) Forecast(
	w nethttp.ResponseWriter,
	r *nethttp.Request,
) {
	forecast, err := h.service.GetForecast(
		r.Context(),
	)
	if err != nil {
		writeWeatherError(w, err)

		return
	}

	writeJSON(
		w,
		nethttp.StatusOK,
		forecast,
	)
}

func writeWeatherError(
	w nethttp.ResponseWriter,
	err error,
) {
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
}
