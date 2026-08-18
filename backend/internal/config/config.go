package config

import "time"

type Config struct {
	Display DisplayConfig
}

type DisplayConfig struct {
	IdleTimeout time.Duration
}

func Default() Config {
	return Config{
		Display: DisplayConfig{
			IdleTimeout: 5 * time.Minute,
		},
	}
}
