package weather

import (
	"context"
	"errors"
	"testing"
)

type stubProvider struct {
	days     int
	forecast []DailyForecast
}

func (p *stubProvider) GetCurrent(
	ctx context.Context,
	latitude float64,
	longitude float64,
	timezone string,
) (CurrentWeather, error) {
	return CurrentWeather{}, nil
}

func (p *stubProvider) GetForecast(
	ctx context.Context,
	latitude float64,
	longitude float64,
	timezone string,
	days int,
) ([]DailyForecast, error) {
	p.days = days

	return p.forecast, nil
}

func TestServiceGetForecastWhenDisabled(t *testing.T) {
	service := NewService(
		false,
		&stubProvider{},
		Location{},
		3,
	)

	_, err := service.GetForecast(context.Background())

	if !errors.Is(err, ErrDisabled) {
		t.Fatalf("expected ErrDisabled, got %v", err)
	}
}

func TestServiceGetForecastUsesConfiguredDays(t *testing.T) {
	provider := &stubProvider{
		forecast: []DailyForecast{{}, {}, {}},
	}

	service := NewService(
		true,
		provider,
		Location{
			Latitude:  -23.55052,
			Longitude: -46.633308,
			Timezone:  "America/Sao_Paulo",
		},
		3,
	)

	forecast, err := service.GetForecast(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if provider.days != 3 {
		t.Errorf(
			"expected the provider to be asked for 3 days, got %d",
			provider.days,
		)
	}

	if len(forecast) != 3 {
		t.Errorf("expected 3 days, got %d", len(forecast))
	}
}

func TestNewServiceClampsForecastDays(t *testing.T) {
	tests := []struct {
		configured int
		expected   int
	}{
		{configured: 0, expected: DefaultForecastDays},
		{configured: -1, expected: DefaultForecastDays},
		{configured: MaxForecastDays + 10, expected: MaxForecastDays},
		{configured: 7, expected: 7},
	}

	for _, test := range tests {
		provider := &stubProvider{}

		service := NewService(
			true,
			provider,
			Location{},
			test.configured,
		)

		if _, err := service.GetForecast(
			context.Background(),
		); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if provider.days != test.expected {
			t.Errorf(
				"configured %d days, expected %d, got %d",
				test.configured,
				test.expected,
				provider.days,
			)
		}
	}
}
