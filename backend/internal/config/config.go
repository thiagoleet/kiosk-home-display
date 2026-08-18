package config

import "time"

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
	Enabled bool
	On      string
	Off     string
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
			Enabled: true,
			On:      "07:00",
			Off:     "23:00",
		},
	}
}
