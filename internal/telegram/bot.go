package telegram

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/drTragger/mykola-miniapp/internal/config"
)

func StartBot(ctx context.Context, cfg config.Config) {
	if cfg.Telegram.Token == "" {
		log.Println("Telegram bot disabled (no token)")
		return
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Printf("telegram bot init failed: %v", err)
		return
	}

	log.Println("Telegram bot started:", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	defer bot.StopReceivingUpdates()

	for {
		select {
		case <-ctx.Done():
			log.Println("Telegram bot stopping...")
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			if update.Message == nil {
				continue
			}
			handleMessage(bot, cfg, update.Message)
		}
	}
}
