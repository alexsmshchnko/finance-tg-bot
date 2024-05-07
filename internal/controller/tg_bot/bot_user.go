package tg_bot

import (
	"context"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotUser struct {
	UserId       int
	ResponseWait bool
	ResponseMsg  tgbotapi.Message
	ResponseCode string
}

var BotUsers map[string]BotUser

func NewBotUser(log *slog.Logger, userName string, id int) {
	if BotUsers == nil {
		BotUsers = make(map[string]BotUser)
	}
	BotUsers[userName] = BotUser{UserId: id}
	log.Info("added to BotUsers", "userName", userName, "len(BotUsers)", len(BotUsers))
}

func responseIsAwaited(userName string) bool {
	return BotUsers[userName].ResponseWait
}

func waitUserResponeStart(log *slog.Logger, userName, respCode string, message tgbotapi.Message) {
	if v, ok := BotUsers[userName]; ok {
		v.ResponseWait = true
		v.ResponseMsg = message
		v.ResponseCode = respCode

		BotUsers[userName] = v
	}
	log.Info("waitUserResponeStart", "userName", userName)
}

func waitUserResponseComplete(log *slog.Logger, userName string) {
	if v, ok := BotUsers[userName]; ok {
		v.ResponseWait = false

		BotUsers[userName] = v
	}
	log.Info("waitUserResponseComplete", "userName", userName)
}

func (b *Bot) checkUser(log *slog.Logger, ctx context.Context, userName string) (f bool) {
	_, f = BotUsers[userName]
	if !f {
		log.Info("checkUser GetUserStatus request", "userName", userName)
		id, active, err := b.accountant.GetUserStatus(ctx, userName)
		if err != nil {
			b.log.Error("checkUser GetUserStatus", "err", err)
		}
		log.Info("checkUser GetUserStatus response", "userName", userName, "active", active)

		if active {
			NewBotUser(log, userName, id)
		}
		f = active
	}
	return
}
