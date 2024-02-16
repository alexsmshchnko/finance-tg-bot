package tg_bot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleSettingCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) {
	split := strings.Split(query.Data, ":")
	switch split[1] {
	case "editCategory":
		b.requestCategoryKeyboardEditor(ctx, 0, query)
	case "cancelSettings":
		b.deleteMsg(query.Message.Chat.ID, query.Message.MessageID)
	}
}

func (b *Bot) handleCategoryKeyboardEditor(ctx context.Context, query *tgbotapi.CallbackQuery) {
	split := strings.Split(query.Data, ":")
	switch split[1] {
	case "addNew":
		b.requestReply(query, "REC_NEWCAT")
	case "goBack":
		b.showSettingsMenu(ctx, nil, query)
	case "newCatRole":
		//save
		direction := 0
		switch split[2] {
		case "debit":
			direction = -1
		case "credit":
			direction = 1
		}
		b.accountant.EditCats(query.Message.Text, direction, true, query.From.UserName)
		//show prev menu
		b.requestCategoryKeyboardEditor(ctx, 0, query)
	default:

	}
}

func (b *Bot) requestCategoryKeyboardEditor(ctx context.Context, page int, q *tgbotapi.CallbackQuery) {
	b.updateMsgText(q.Message.Chat.ID, q.Message.MessageID, "Настройки категорий")
	cats, err := b.accountant.GetCats(ctx, q.From.UserName)
	if err != nil {
		return
	}

	options := make([][]string, 0, len(cats)+1)
	for _, v := range cats {
		options = append(options, []string{v, PREFIX_SETCATEGORY + ":" + v})
	}
	options = append(options, []string{EMOJI_ADD, PREFIX_SETCATEGORY + ":addNew"})
	mrkp := newKeyboardForm()
	mrkp.setOptions(options)
	mrkp.addNavigationControl(page, []string{EMOJI_HOOK_BACK, PREFIX_SETCATEGORY + ":goBack"}, nil)
	resMrkp, err := mrkp.getMarkup()
	if err != nil {
		fmt.Println(err)
		return
	}

	msg := tgbotapi.NewEditMessageReplyMarkup(q.Message.Chat.ID, q.Message.MessageID, *resMrkp)
	b.api.Send(msg)
}
