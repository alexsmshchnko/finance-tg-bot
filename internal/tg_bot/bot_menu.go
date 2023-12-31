package tg_bot

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func initCommands() (conf tgbotapi.SetMyCommandsConfig) {
	commands := []tgbotapi.BotCommand{
		{
			Command:     "/start",
			Description: "Запустить бота",
		},
		{
			Command:     "/help",
			Description: "Помощь",
		},
		{
			Command:     "/sync",
			Description: "Загрузить историю с облака",
		},
		{
			Command:     "/push",
			Description: "Экспорт в облако",
		},
		{
			Command:     "/settings",
			Description: "Настройки",
		},
	}

	conf = tgbotapi.NewSetMyCommands(commands...)

	return
}

func (b *Bot) processCommand(ctx context.Context, u *tgbotapi.Update) (err error) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "")

	switch u.Message.Command() {
	case "help":
		msg.Text = "I understand /sayhi and /status."
	case "start":
		msg.Text = "Hi :)"
	case "sync":
		b.syncCmd(ctx, u)
	default:
		msg.Text = "I don't know that command"
	}
	if err != nil {
		log.Println(err)
		return
	}
	_, err = b.api.Send(msg)

	return
}
