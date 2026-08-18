package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Display   DisplayConfig
	Idle      IdleConfig
	Scheduler SchedulerConfig
}

type DisplayConfig struct {
	Mode       string
	Brightness int
}

type IdleConfig struct {
	Enabled bool
	Timeout time.Duration
}

type SchedulerConfig struct {
	Enabled  bool
	On       string
	Off      string
	Timezone string
}

func Default() Config {
	return Config{
		Display: DisplayConfig{
			Mode:       "virtual",
			Brightness: 100,
		},
		Idle: IdleConfig{
			Enabled: true,
			Timeout: 5 * time.Minute,
		},
		Scheduler: SchedulerConfig{
			Enabled:  true,
			On:       "07:00",
			Off:      "23:00",
			Timezone: "America/Sao_Paulo",
		},
	}
}

func (c Config) Validate() error {
	if c.Display.Mode != "virtual" && c.Display.Mode != "linux" {
		return fmt.Errorf(
			"invalid display mode: %q",
			c.Display.Mode,
		)
	}

	if c.Display.Brightness < 0 || c.Display.Brightness > 100 {
		return fmt.Errorf(
			"display brightness must be between 0 and 100",
		)
	}

	if c.Idle.Enabled && c.Idle.Timeout <= 0 {
		return fmt.Errorf(
			"idle timeout must be greater than zero",
		)
	}

	if c.Scheduler.Enabled {
		if c.Scheduler.On == "" {
			return fmt.Errorf(
				"schedule on time cannot be empty",
			)
		}

		if c.Scheduler.Off == "" {
			return fmt.Errorf(
				"schedule off time cannot be empty",
			)
		}

		if c.Scheduler.Timezone == "" {
			return fmt.Errorf(
				"scheduler timezone cannot be empty",
			)
		}

		if _, err := time.LoadLocation(c.Scheduler.Timezone); err != nil {
			return fmt.Errorf(
				"invalid scheduler timezone %q: %w",
				c.Scheduler.Timezone,
				err,
			)
		}
	}

	return nil
}

func Load() (Config, error) {
	config := Default()

	if value := os.Getenv("DISPLAY_MODE"); value != "" {
		config.Display.Mode = value
	}

	if value := os.Getenv("DISPLAY_BRIGHTNESS"); value != "" {
		brightness, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf(
				"invalid DISPLAY_BRIGHTNESS: %w",
				err,
			)
		}

		config.Display.Brightness = brightness
	}

	if value := os.Getenv("IDLE_ENABLED"); value != "" {
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf(
				"invalid IDLE_ENABLED: %w",
				err,
			)
		}

		config.Idle.Enabled = enabled
	}

	if value := os.Getenv("IDLE_TIMEOUT"); value != "" {
		timeout, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf(
				"invalid IDLE_TIMEOUT: %w",
				err,
			)
		}

		config.Idle.Timeout = timeout
	}

	if value := os.Getenv("SCHEDULE_ENABLED"); value != "" {
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf(
				"invalid SCHEDULE_ENABLED: %w",
				err,
			)
		}

		config.Scheduler.Enabled = enabled
	}

	if value := os.Getenv("SCHEDULE_ON"); value != "" {
		config.Scheduler.On = value
	}

	if value := os.Getenv("SCHEDULE_OFF"); value != "" {
		config.Scheduler.Off = value
	}

	if value := os.Getenv("TIMEZONE"); value != "" {
		config.Scheduler.Timezone = value
	}

	if err := config.Validate(); err != nil {
		return Config{}, err
	}

	return config, nil
}
