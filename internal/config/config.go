package config

import "os"

type Config struct {
	AppAddr string
}

func Load() Config {
	appAddr := os.Getenv("APP_ADDR")
	if appAddr == "" {
		appAddr = ":8090"
	}

	return Config{
		AppAddr: appAddr,
	}
}
