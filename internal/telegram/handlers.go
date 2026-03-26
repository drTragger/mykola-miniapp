package telegram

import (
	"log"
	"os/exec"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/drTragger/mykola-miniapp/internal/config"
)

func handleMessage(bot *tgbotapi.BotAPI, cfg config.Config, msg *tgbotapi.Message) {
	if !msg.IsCommand() {
		return
	}

	if !IsAdminUser(cfg, msg.From.ID) {
		send(bot, msg.Chat.ID, "⛔ Немає доступу")
		return
	}

	switch msg.Command() {
	case "poweroff":
		send(bot, msg.Chat.ID, "⚠️ Вимикаю систему...")
		go runCommand("sudo", "systemctl", "poweroff")

	case "reboot":
		send(bot, msg.Chat.ID, "🔄 Перезавантажую...")
		go runCommand("sudo", "systemctl", "reboot")

	case "ping":
		send(bot, msg.Chat.ID, "pong 🏓")

	default:
		send(bot, msg.Chat.ID, "❓ Невідома команда")
	}
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, _ = bot.Send(msg)
}

func runCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("command error: %v, output: %s", err, strings.TrimSpace(string(output)))
		return
	}

	log.Printf("command success: %s %s", name, strings.Join(args, " "))
}
