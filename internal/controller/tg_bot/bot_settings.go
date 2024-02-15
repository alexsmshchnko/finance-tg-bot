package tg_bot

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleSettingCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) {
	split := strings.Split(query.Data, ":")
	switch split[1] {
	case "editCategory":
		b.editCategoryKeyboard(ctx, query)
	case "cancelSettings":
		b.cancelSettings(query)
	}
}
func (b *Bot) cancelSettings(q *tgbotapi.CallbackQuery) {
	b.deleteMsg(q.Message.Chat.ID, q.Message.MessageID)
}

func (b *Bot) editCategoryKeyboard(ctx context.Context, q *tgbotapi.CallbackQuery) {
	cats, err := b.accountant.GetCats(ctx, q.From.UserName)
	if err != nil {
		return
	}
	cats = addButtonToSlice(cats, EMOJI_ADD+" (добавить)")

	mrkp := getPagedListInlineKeyboard(cats, 0, PREFIX_SETCATEGORY, "")

	msg := tgbotapi.NewEditMessageReplyMarkup(q.Message.Chat.ID, q.Message.MessageID, *mrkp)
	b.api.Send(msg)
}
