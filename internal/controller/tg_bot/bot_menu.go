package tg_bot

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func initCommands() (conf tgbotapi.SetMyCommandsConfig) {
	conf = tgbotapi.NewSetMyCommands(
		[]tgbotapi.BotCommand{
			{Command: "/report", Description: "Заказать отчет"},
			{Command: "/settings", Description: "Настройки"},
			{Command: "/push", Description: "Экспорт в облако"},
		}...,
	)

	return
}

func (b *Bot) processCommand(ctx context.Context, u *tgbotapi.Update) {
	switch u.Message.Command() {
	case "start":
		b.api.Send(tgbotapi.NewMessage(u.Message.Chat.ID, "Hi :)"))
	case "sync":
		b.syncCmd(ctx, u)
	case "push":
		b.pushCmd(ctx, u)
	case "report":
		b.showReportMenu(u)
	case "settings":
		b.showSettingsMenu(u, nil)
	default:
		b.api.Send(tgbotapi.NewMessage(u.Message.Chat.ID, "I don't know that command"))
	}
}

func (b *Bot) runSyncLoad(ctx context.Context, userName string) (msg string) {
	msg = EMOJI_THUMB_UP
	err := b.accountant.MigrateFromCloud(ctx, userName)
	if err != nil {
		b.log.Error("runSyncLoad", "err", err)
		msg = EMOJI_THUMB_DOWN
	}

	return
}

func (b *Bot) runSyncUpload(ctx context.Context, userName string) (msg string) {
	msg = EMOJI_THUMB_UP
	err := b.accountant.PushToCloud(ctx, userName)
	if err != nil {
		b.log.Error("runSyncUpload", "err", err)
		msg = EMOJI_THUMB_DOWN
	}

	return
}

func (b *Bot) syncCmd(ctx context.Context, u *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, EMOJI_ROCKET)
	msg.ReplyToMessageID = u.Message.MessageID
	startMsg, _ := b.api.Send(msg) //start sync

	msg.Text = b.runSyncLoad(ctx, u.Message.Chat.UserName)
	b.updateMsgText(startMsg.Chat.ID, startMsg.MessageID, msg.Text)

	go func(sec time.Duration) {
		time.Sleep(sec * time.Second)

		b.deleteMsg(startMsg.Chat.ID, u.Message.MessageID)
		b.deleteMsg(startMsg.Chat.ID, startMsg.MessageID)
	}(4)
}

func (b *Bot) pushCmd(ctx context.Context, u *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, EMOJI_ROCKET)
	msg.ReplyToMessageID = u.Message.MessageID
	startMsg, _ := b.api.Send(msg) //start sync

	msg.Text = b.runSyncUpload(ctx, u.Message.Chat.UserName)
	b.updateMsgText(startMsg.Chat.ID, startMsg.MessageID, msg.Text)

	go func(sec time.Duration) {
		time.Sleep(sec * time.Second)

		b.deleteMsg(startMsg.Chat.ID, u.Message.MessageID)
		b.deleteMsg(startMsg.Chat.ID, startMsg.MessageID)
	}(4)
}

func (b *Bot) showReportMenu(u *tgbotapi.Update) {
	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID)
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "Отчеты")
	msg.ReplyMarkup = getReportKeyboard()
	b.api.Send(msg)
}

func (b *Bot) showSettingsMenu(u *tgbotapi.Update, q *tgbotapi.CallbackQuery) {
	if q == nil {
		b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID)
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, "Настройки")
		msg.ReplyMarkup = getSettingsKeyboard()
		b.api.Send(msg)
	} else {
		b.updateMsgText(q.Message.Chat.ID, q.Message.MessageID, "Настройки")
		b.api.Send(tgbotapi.NewEditMessageReplyMarkup(q.Message.Chat.ID, q.Message.MessageID, getSettingsKeyboard()))
	}
}
