package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/drTragger/mykola-miniapp/internal/config"
)

func StartBot(cfg config.Config) {
	if cfg.Telegram.Token == "" {
		log.Println("Telegram bot disabled (no token)")
		return
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Telegram bot started:", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		handleMessage(bot, cfg, update.Message)
	}
}
