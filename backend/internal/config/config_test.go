package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := Default()

	if config.Display.Mode != "virtual" {
		t.Fatalf(
			"expected virtual display mode, got %q",
			config.Display.Mode,
		)
	}

	if config.Display.Brightness != 100 {
		t.Fatalf(
			"expected brightness 100, got %d",
			config.Display.Brightness,
		)
	}

	if !config.Idle.Enabled {
		t.Fatal("expected idle to be enabled")
	}

	if config.Idle.Timeout != 5*time.Minute {
		t.Fatalf(
			"expected idle timeout of 5 minutes, got %s",
			config.Idle.Timeout,
		)
	}

	if !config.Scheduler.Enabled {
		t.Fatal("expected scheduler to be enabled")
	}

	if config.Scheduler.On != "07:00" {
		t.Fatalf(
			"expected schedule on at 07:00, got %q",
			config.Scheduler.On,
		)
	}

	if config.Scheduler.Off != "23:00" {
		t.Fatalf(
			"expected schedule off at 23:00, got %q",
			config.Scheduler.Off,
		)
	}

	if config.Scheduler.Timezone != "America/Sao_Paulo" {
		t.Fatalf(
			"expected timezone America/Sao_Paulo, got %q",
			config.Scheduler.Timezone,
		)
	}

	if config.Activity.LifeSpan != 7*24*time.Hour {
		t.Fatalf(
			"expected activity life span of 7 days, got %s",
			config.Activity.LifeSpan,
		)
	}

	if config.Weather.Enabled {
		t.Fatal("expected weather to be disabled")
	}

	if config.Weather.OpenMeteoAPIURL != "https://api.open-meteo.com/v1/forecast" {
		t.Fatalf(
			"expected default Open-Meteo API URL, got %q",
			config.Weather.OpenMeteoAPIURL,
		)
	}

	if config.Weather.Timezone != "America/Sao_Paulo" {
		t.Fatalf(
			"expected weather timezone America/Sao_Paulo, got %q",
			config.Weather.Timezone,
		)
	}

	if config.Weather.CacheTTL != 5*time.Minute {
		t.Fatalf(
			"expected weather cache ttl of 5 minutes, got %s",
			config.Weather.CacheTTL,
		)
	}
}

func TestLoadOverridesDefaults(t *testing.T) {
	t.Setenv("DISPLAY_MODE", "linux")
	t.Setenv("DISPLAY_BRIGHTNESS", "80")

	t.Setenv("IDLE_ENABLED", "false")
	t.Setenv("IDLE_TIMEOUT", "10m")

	t.Setenv("SCHEDULE_ENABLED", "false")
	t.Setenv("SCHEDULE_ON", "08:00")
	t.Setenv("SCHEDULE_OFF", "22:00")
	t.Setenv("TIMEZONE", "America/New_York")
	t.Setenv("ACTIVITY_LIFE_SPAN", "48h")
	t.Setenv("WEATHER_ENABLED", "true")
	t.Setenv("OPEN_METEO_API_URL", "https://weather.example.com/forecast")
	t.Setenv("WEATHER_LATITUDE", "40.7128")
	t.Setenv("WEATHER_LONGITUDE", "-74.0060")
	t.Setenv("WEATHER_TIMEZONE", "America/New_York")
	t.Setenv("WEATHER_CACHE_TTL", "10m")

	config, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.Display.Mode != "linux" {
		t.Fatalf(
			"expected linux mode, got %q",
			config.Display.Mode,
		)
	}

	if config.Display.Brightness != 80 {
		t.Fatalf(
			"expected brightness 80, got %d",
			config.Display.Brightness,
		)
	}

	if config.Idle.Enabled {
		t.Fatal("expected idle to be disabled")
	}

	if config.Idle.Timeout != 10*time.Minute {
		t.Fatalf(
			"expected 10 minute timeout, got %s",
			config.Idle.Timeout,
		)
	}

	if config.Scheduler.Enabled {
		t.Fatal("expected scheduler to be disabled")
	}

	if config.Scheduler.Timezone != "America/New_York" {
		t.Fatalf(
			"expected timezone America/New_York, got %q",
			config.Scheduler.Timezone,
		)
	}

	if config.Activity.LifeSpan != 48*time.Hour {
		t.Fatalf(
			"expected 48h activity life span, got %s",
			config.Activity.LifeSpan,
		)
	}

	if !config.Weather.Enabled {
		t.Fatal("expected weather to be enabled")
	}

	if config.Weather.OpenMeteoAPIURL != "https://weather.example.com/forecast" {
		t.Fatalf(
			"expected configured Open-Meteo API URL, got %q",
			config.Weather.OpenMeteoAPIURL,
		)
	}

	if config.Weather.Latitude != 40.7128 {
		t.Fatalf(
			"expected weather latitude 40.7128, got %f",
			config.Weather.Latitude,
		)
	}

	if config.Weather.Longitude != -74.0060 {
		t.Fatalf(
			"expected weather longitude -74.0060, got %f",
			config.Weather.Longitude,
		)
	}

	if config.Weather.Timezone != "America/New_York" {
		t.Fatalf(
			"expected weather timezone America/New_York, got %q",
			config.Weather.Timezone,
		)
	}

	if config.Weather.CacheTTL != 10*time.Minute {
		t.Fatalf(
			"expected weather cache ttl 10m, got %s",
			config.Weather.CacheTTL,
		)
	}
}

func TestLoadRejectsInvalidDisplayMode(t *testing.T) {
	t.Setenv("DISPLAY_MODE", "invalid")

	_, err := Load()

	if err == nil {
		t.Fatal("expected configuration error")
	}
}

func TestLoadRejectsInvalidBrightness(t *testing.T) {
	t.Setenv("DISPLAY_BRIGHTNESS", "150")

	config, err := Load()

	if err == nil {
		t.Fatal("expected configuration error")
	}

	if config.Display.Brightness != 0 {
		t.Fatal("expected empty config on error")
	}
}

func TestLoadRejectsInvalidTimezone(t *testing.T) {
	t.Setenv("TIMEZONE", "Invalid/Timezone")

	_, err := Load()

	if err == nil {
		t.Fatal("expected configuration error")
	}
}

func TestLoadRejectsInvalidActivityLifeSpan(t *testing.T) {
	t.Setenv("ACTIVITY_LIFE_SPAN", "not-a-duration")

	_, err := Load()

	if err == nil {
		t.Fatal("expected configuration error")
	}
}

func TestLoadRejectsZeroActivityLifeSpan(t *testing.T) {
	t.Setenv("ACTIVITY_LIFE_SPAN", "0s")

	_, err := Load()

	if err == nil {
		t.Fatal("expected configuration error")
	}
}

func TestLoadRejectsInvalidWeatherLatitude(t *testing.T) {
	t.Setenv("WEATHER_ENABLED", "true")
	t.Setenv("WEATHER_LATITUDE", "invalid")

	_, err := Load()

	if err == nil {
		t.Fatal("expected configuration error")
	}
}

func TestLoadRejectsWeatherLatitudeOutOfRange(t *testing.T) {
	t.Setenv("WEATHER_ENABLED", "true")
	t.Setenv("WEATHER_LATITUDE", "100")

	_, err := Load()

	if err == nil {
		t.Fatal("expected configuration error")
	}
}

func TestLoadRejectsInvalidWeatherTimezone(t *testing.T) {
	t.Setenv("WEATHER_ENABLED", "true")
	t.Setenv("WEATHER_TIMEZONE", "Invalid/Timezone")

	_, err := Load()

	if err == nil {
		t.Fatal("expected configuration error")
	}
}

func TestLoadRejectsZeroWeatherCacheTTL(t *testing.T) {
	t.Setenv("WEATHER_ENABLED", "true")
	t.Setenv("WEATHER_CACHE_TTL", "0s")

	_, err := Load()

	if err == nil {
		t.Fatal("expected configuration error")
	}
}
