package main

import (
	"log"
	"time"

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

	return
}

func syncCmd(u *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "\U0001f680")
	msg.ReplyToMessageID = u.Message.MessageID
	startMsg, _ := gBot.Send(msg) //start sync

	msg.Text = runSync(u.Message.Chat.UserName)
	updateMsgText(startMsg.Chat.ID, startMsg.MessageID, runSync(u.Message.Chat.UserName))

	go func(sec time.Duration) {
		time.Sleep(sec * time.Second)

		deleteMsg(startMsg.Chat.ID, u.Message.MessageID)
		deleteMsg(startMsg.Chat.ID, startMsg.MessageID)
	}(4)
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
