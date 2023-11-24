package main

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func runSync(userName string) (msg string) {
	msg = "\U0001f44d"
	err := sync.MigrateFromCloud(ctx, userName)
	if err != nil {
		log.Println(err)
		msg = "\U0001f44e"
	}

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
