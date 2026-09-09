package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const dailyForecastPayload = `{
	"daily": {
		"time": ["2026-09-09", "2026-09-10", "2026-09-11"],
		"weather_code": [0, 61, 95],
		"temperature_2m_max": [28.4, 24.1, 21.8],
		"temperature_2m_min": [17.2, 16.5, 15.9],
		"precipitation_probability_max": [0, 80, 95]
	}
}`

func TestOpenMeteoProviderGetForecast(t *testing.T) {
	var query url.Values

	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			query = r.URL.Query()

			_, _ = w.Write([]byte(dailyForecastPayload))
		}),
	)

	defer server.Close()

	provider := NewOpenMeteoProvider(
		server.Client(),
		server.URL,
	)

	forecast, err := provider.GetForecast(
		context.Background(),
		-23.55052,
		-46.633308,
		"America/Sao_Paulo",
		3,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if query.Get("forecast_days") != "3" {
		t.Errorf(
			"expected forecast_days 3, got %q",
			query.Get("forecast_days"),
		)
	}

	if query.Get("timezone") != "America/Sao_Paulo" {
		t.Errorf(
			"expected timezone America/Sao_Paulo, got %q",
			query.Get("timezone"),
		)
	}

	if query.Get("daily") == "" {
		t.Error("expected the daily parameter to be requested")
	}

	if len(forecast) != 3 {
		t.Fatalf("expected 3 days, got %d", len(forecast))
	}

	first := forecast[0]

	if first.Date.Format("2006-01-02") != "2026-09-09" {
		t.Errorf(
			"expected first day 2026-09-09, got %s",
			first.Date.Format("2006-01-02"),
		)
	}

	if first.Date.Location().String() != "America/Sao_Paulo" {
		t.Errorf(
			"expected the date in the configured timezone, got %s",
			first.Date.Location(),
		)
	}

	if first.TemperatureMax != 28.4 || first.TemperatureMin != 17.2 {
		t.Errorf(
			"expected 17.2/28.4, got %f/%f",
			first.TemperatureMin,
			first.TemperatureMax,
		)
	}

	if first.Condition != ConditionClear {
		t.Errorf(
			"expected condition clear, got %q",
			first.Condition,
		)
	}

	if forecast[1].Condition != ConditionRain {
		t.Errorf(
			"expected condition rain, got %q",
			forecast[1].Condition,
		)
	}

	if forecast[2].PrecipitationProbability != 95 {
		t.Errorf(
			"expected precipitation probability 95, got %f",
			forecast[2].PrecipitationProbability,
		)
	}
}

func TestOpenMeteoProviderGetForecastWithoutPrecipitation(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			_, _ = w.Write([]byte(`{
				"daily": {
					"time": ["2026-09-09"],
					"weather_code": [3],
					"temperature_2m_max": [20.0],
					"temperature_2m_min": [12.0]
				}
			}`))
		}),
	)

	defer server.Close()

	provider := NewOpenMeteoProvider(
		server.Client(),
		server.URL,
	)

	forecast, err := provider.GetForecast(
		context.Background(),
		0,
		0,
		"UTC",
		1,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(forecast) != 1 {
		t.Fatalf("expected 1 day, got %d", len(forecast))
	}

	if forecast[0].PrecipitationProbability != 0 {
		t.Errorf(
			"expected precipitation probability 0, got %f",
			forecast[0].PrecipitationProbability,
		)
	}
}

func TestOpenMeteoProviderGetForecastRejectsIncompleteSeries(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			_, _ = w.Write([]byte(`{
				"daily": {
					"time": ["2026-09-09", "2026-09-10"],
					"weather_code": [0],
					"temperature_2m_max": [20.0],
					"temperature_2m_min": [12.0]
				}
			}`))
		}),
	)

	defer server.Close()

	provider := NewOpenMeteoProvider(
		server.Client(),
		server.URL,
	)

	if _, err := provider.GetForecast(
		context.Background(),
		0,
		0,
		"UTC",
		2,
	); err == nil {
		t.Fatal("expected an error for a short series")
	}
}

func TestOpenMeteoProviderGetForecastRejectsInvalidDays(t *testing.T) {
	provider := NewOpenMeteoProvider(nil, "")

	for _, days := range []int{0, -1, MaxForecastDays + 1} {
		if _, err := provider.GetForecast(
			context.Background(),
			0,
			0,
			"UTC",
			days,
		); err == nil {
			t.Errorf(
				"expected an error for %d days",
				days,
			)
		}
	}
}

func TestOpenMeteoProviderGetForecastFailsOnErrorStatus(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusInternalServerError)
		}),
	)

	defer server.Close()

	provider := NewOpenMeteoProvider(
		server.Client(),
		server.URL,
	)

	if _, err := provider.GetForecast(
		context.Background(),
		0,
		0,
		"UTC",
		3,
	); err == nil {
		t.Fatal("expected an error for a non-200 response")
	}
}
