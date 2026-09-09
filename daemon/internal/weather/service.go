package weather

import (
	"context"
	"errors"
)

var ErrDisabled = errors.New("weather forecast is not enabled")

// DefaultForecastDays is how far the outlook reaches when nothing else
// is configured.
const DefaultForecastDays = 5

type Location struct {
	Latitude  float64
	Longitude float64
	Timezone  string
}

type Service struct {
	provider     Provider
	location     Location
	forecastDays int
	enabled      bool
}

func NewService(
	enabled bool,
	provider Provider,
	location Location,
	forecastDays int,
) *Service {
	if forecastDays < MinForecastDays {
		forecastDays = DefaultForecastDays
	}

	if forecastDays > MaxForecastDays {
		forecastDays = MaxForecastDays
	}

	return &Service{
		provider:     provider,
		location:     location,
		forecastDays: forecastDays,
		enabled:      enabled,
	}
}

func (s *Service) GetCurrent(
	ctx context.Context,
) (CurrentWeather, error) {
	if !s.enabled {
		return CurrentWeather{}, ErrDisabled
	}

	return s.provider.GetCurrent(
		ctx,
		s.location.Latitude,
		s.location.Longitude,
		s.location.Timezone,
	)
}

// GetForecast returns the configured number of days, today included.
func (s *Service) GetForecast(
	ctx context.Context,
) ([]DailyForecast, error) {
	if !s.enabled {
		return nil, ErrDisabled
	}

	return s.provider.GetForecast(
		ctx,
		s.location.Latitude,
		s.location.Longitude,
		s.location.Timezone,
		s.forecastDays,
	)
}
