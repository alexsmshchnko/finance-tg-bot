package tg_bot

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotUser struct {
	ResponseWait  bool
	ResponseMsg   tgbotapi.Message
	ResponseCode  string
	FinCategories []string
}

var BotUsers map[string]BotUser

func NewBotUser(userName string) {
	BotUsers = map[string]BotUser{userName: {}}
	log.Printf("BotUsers: %v\n", BotUsers)
}

func responseIsAwaited(userName string) bool {
	return BotUsers[userName].ResponseWait
}

func waitUserResponeStart(userName, respCode string, message tgbotapi.Message) {
	if _, ok := BotUsers[userName]; ok {
		BotUsers[userName] = BotUser{
			ResponseWait: true,
			ResponseMsg:  message,
			ResponseCode: respCode,
		}
	}
	log.Printf("ResponseWaitStart: %v\n", BotUsers[userName])
}

func waitUserResponseComplete(userName string) {
	if _, ok := BotUsers[userName]; ok {
		BotUsers[userName] = BotUser{
			ResponseWait: false,
		}
	}
	log.Printf("ResponseWaitStop: %v\n", BotUsers[userName])
}

func (b *Bot) checkUser(ctx context.Context, userName string) bool {
	_, f := BotUsers[userName]
	if !f {
		active, err := b.accountant.GetUserStatus(ctx, userName)
		if err != nil {
			log.Println(err)
		}
		if active {
			NewBotUser(userName)
		}
		return active
	}
	return f
}
