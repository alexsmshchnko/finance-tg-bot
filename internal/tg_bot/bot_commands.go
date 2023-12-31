package tg_bot

import (
	"context"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) runSync(ctx context.Context, userName string) (msg string) {
	msg = EMOJI_THUMB_UP
	err := b.synchronizer.MigrateFromCloud(ctx, userName)
	if err != nil {
		log.Println(err)
		msg = EMOJI_THUMB_DOWN
	}

	return
}

func (b *Bot) syncCmd(ctx context.Context, u *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, EMOJI_ROCKET)
	msg.ReplyToMessageID = u.Message.MessageID
	startMsg, _ := b.api.Send(msg) //start sync

	msg.Text = b.runSync(ctx, u.Message.Chat.UserName)
	b.updateMsgText(startMsg.Chat.ID, startMsg.MessageID, msg.Text)

	go func(sec time.Duration) {
		time.Sleep(sec * time.Second)

		b.deleteMsg(startMsg.Chat.ID, u.Message.MessageID)
		b.deleteMsg(startMsg.Chat.ID, startMsg.MessageID)
	}(4)
}
