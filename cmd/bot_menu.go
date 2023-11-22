package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// type (
// 	commandEntity struct {
// 		key    string
// 		desc   string
// 		action func(upd tgbotapi.Update)
// 	}
// )

// const (
// 	StartCmdKey    = "start"
// 	HelpCmdKey     = "help"
// 	SyncCmdKey     = "sync"
// 	SettingsCmdKey = "settings"
// )

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
			Description: "Экспорт облако",
		},
		{
			Command:     "/settings",
			Description: "Настройки",
		},
	}

	conf = tgbotapi.NewSetMyCommands(commands...)

	return
}

func processCommand(u *tgbotapi.Update) (err error) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "")

	// Extract the command from the Message.
	switch u.Message.Command() {
	case "help":
		msg.Text = "I understand /sayhi and /status."
	case "start":
		msg.Text = "Hi :)"
	case "sync":
		syncCmd(u)
	default:
		msg.Text = "I don't know that command"
	}
	if err != nil {
		log.Println(err)
		return
	}
	_, err = gBot.Send(msg)

	return
}
