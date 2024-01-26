package tg_bot

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func initCommands() (conf tgbotapi.SetMyCommandsConfig) {
	commands := []tgbotapi.BotCommand{
		{
			Command:     "/settings",
			Description: "Настройки",
		},
		{
			Command:     "/push",
			Description: "Экспорт в облако",
		},
		{
			Command:     "/sync",
			Description: "Загрузить историю с облака",
		},
	}

	conf = tgbotapi.NewSetMyCommands(commands...)

	return
}

func (b *Bot) processCommand(ctx context.Context, u *tgbotapi.Update) (err error) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "")

	switch u.Message.Command() {
	case "start":
		msg.Text = "Hi :)"
	case "sync":
		b.syncCmd(ctx, u)
	case "push":
		b.pushCmd(ctx, u)
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
