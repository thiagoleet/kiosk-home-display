package weather

import "context"

type Location struct {
	Latitude  float64
	Longitude float64
	Timezone  string
}

type Service struct {
	provider Provider
	location Location
}

func NewService(
	provider Provider,
	location Location,
) *Service {
	return &Service{
		provider: provider,
		location: location,
	}
}

func (s *Service) GetCurrent(
	ctx context.Context,
) (CurrentWeather, error) {
	return s.provider.GetCurrent(
		ctx,
		s.location.Latitude,
		s.location.Longitude,
		s.location.Timezone,
	)
}
