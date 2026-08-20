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
