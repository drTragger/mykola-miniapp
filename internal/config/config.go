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

	QBittorrent struct {
		BaseURL  string `toml:"base_url"`
		Username string `toml:"username"`
		Password string `toml:"password"`
	} `toml:"qbittorrent"`
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

	if cfg.QBittorrent.BaseURL == "" {
		cfg.QBittorrent.BaseURL = "http://127.0.0.1:8080"
	}

	return cfg
}
