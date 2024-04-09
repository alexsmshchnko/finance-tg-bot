package tg_bot

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotUser struct {
	UserId       int
	ResponseWait bool
	ResponseMsg  tgbotapi.Message
	ResponseCode string
}

var BotUsers map[string]BotUser

func NewBotUser(userName string, id int) {
	if BotUsers == nil {
		BotUsers = make(map[string]BotUser)
	}
	BotUsers[userName] = BotUser{UserId: id}
	log.Printf("added BotUsers: %v\n", BotUsers)
}

func responseIsAwaited(userName string) bool {
	return BotUsers[userName].ResponseWait
}

func waitUserResponeStart(userName, respCode string, message tgbotapi.Message) {
	if v, ok := BotUsers[userName]; ok {
		v.ResponseWait = true
		v.ResponseMsg = message
		v.ResponseCode = respCode

		BotUsers[userName] = v
	}
	log.Printf("ResponseWaitStart: %v\n", BotUsers[userName])
}

func waitUserResponseComplete(userName string) {
	if v, ok := BotUsers[userName]; ok {
		v.ResponseWait = false

		BotUsers[userName] = v
	}
	log.Printf("ResponseWaitStop: %v\n", BotUsers[userName])
}

func (b *Bot) checkUser(ctx context.Context, userName string) bool {
	_, f := BotUsers[userName]
	log.Printf("Bot.checkUser cache found: %v\n", f)
	if !f {
		log.Printf("b.accountant.GetUserStatus request: %s\n", userName)
		id, active, err := b.accountant.GetUserStatus(ctx, userName)
		if err != nil {
			log.Println(err)
		}
		log.Printf("b.accountant.GetUserStatus response: %s : %v\n", userName, active)

		if active {
			NewBotUser(userName, id)
		}
		return active
	}
	return f
}
