package telegram

import "github.com/drTragger/mykola-miniapp/internal/config"

func IsAdminUser(cfg config.Config, userID int64) bool {
	for _, id := range cfg.Telegram.AdminIDs {
		if id == userID {
			return true
		}
	}
	return false
}
