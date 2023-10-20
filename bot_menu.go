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
			Description: "Синхронизация с диском",
		},
		{
			Command:     "/settings",
			Description: "Настройки",
		},
	}

	conf = tgbotapi.NewSetMyCommands(commands...)

	// tgCommands := make([]tgbotapi.BotCommand, 0, len(commands))
	// for _, cmd := range commands {
	// 	b.commands[cmd.key] = cmd
	// 	tgCommands = append(tgCommands, tgbotapi.BotCommand{
	// 		Command:     "/" + string(cmd.key),
	// 		Description: cmd.desc,
	// 	})
	// }

	// conf = tgbotapi.NewSetMyCommands(tgCommands...)
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
		msg.Text = "\U0001f680" //start sync
		gBot.Send(msg)
		msg.Text = runSync(u.Message.Chat.UserName)
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
