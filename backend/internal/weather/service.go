package weather

import (
	"context"
	"errors"
)

var ErrDisabled = errors.New("weather forecast is not enabled")

type Location struct {
	Latitude  float64
	Longitude float64
	Timezone  string
}

type Service struct {
	provider Provider
	location Location
	enabled  bool
}

func NewService(
	enabled bool,
	provider Provider,
	location Location,
) *Service {
	return &Service{
		provider: provider,
		location: location,
		enabled:  enabled,
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
