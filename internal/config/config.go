package config

import (
	"log"
	"sync"

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

	Debug struct {
		AllowDevBypass bool `toml:"allow_dev_bypass"`
	}
}

var (
	loadOnce sync.Once
	cached   Config
)

func Load() Config {
	loadOnce.Do(func() {
		if _, err := toml.DecodeFile("config.toml", &cached); err != nil {
			log.Fatal("failed to load config.toml: ", err)
		}

		if cached.App.Addr == "" {
			cached.App.Addr = ":8090"
		}

		if cached.QBittorrent.BaseURL == "" {
			cached.QBittorrent.BaseURL = "http://127.0.0.1:8080"
		}
	})

	return cached
}
