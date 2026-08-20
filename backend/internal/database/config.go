package database

type Config struct {
	Path string
}

func DefaultConfig() Config {
	return Config{
		Path: "./data/kiosk.db",
	}
}
