package config

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Config struct {
	App struct {
		Addr string `toml:"addr"`
	} `toml:"app"`

	Telegram struct {
		Token    string  `toml:"token"`
		AdminIDs []int64 `toml:"admin_ids"`
	} `toml:"telegram"`
}

func Load() Config {
	var cfg Config

	_, err := toml.DecodeFile("config.toml", &cfg)
	if err != nil {
		log.Fatal("failed to load config.toml:", err)
	}

	if cfg.App.Addr == "" {
		cfg.App.Addr = ":8090"
	}

	return cfg
}
