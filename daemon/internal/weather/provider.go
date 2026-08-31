package weather

import "context"

type Provider interface {
	GetCurrent(
		ctx context.Context,
		latitude float64,
		longitude float64,
		timezone string,
	) (CurrentWeather, error)
}
