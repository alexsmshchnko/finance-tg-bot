package tg_bot

import (
	"context"
	"finance-tg-bot/internal/entity"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleSettingCallbackQuery(ctx context.Context, q *tgbotapi.CallbackQuery) {
	split := strings.Split(q.Data, ":")
	switch split[1] {
	case "editCategory":
		requestCategoriesKeyboardEditor(b, ctx, 0, &userChat{q.Message.Chat.ID, q.Message.MessageID, q.From.UserName})
	case "cancelSettings":
		b.deleteMsg(q.Message.Chat.ID, q.Message.MessageID)
	}
}

func (b *Bot) handleCategoryKeyboardEditor(ctx context.Context, q *tgbotapi.CallbackQuery) {
	split := strings.Split(q.Data, ":")
	switch split[1] {
	case "addNewCat":
		b.requestReply(q, "REC_NEWCAT")
	case "backToSettings":
		b.showSettingsMenu(ctx, nil, q)
	case "backToCategories":
		requestCategoriesKeyboardEditor(b, ctx, 0, &userChat{q.Message.Chat.ID, q.Message.MessageID, q.From.UserName})
	case "new":
		str, _ := strings.CutPrefix(q.Message.Text, "Тип траты для ")
		cat := &entity.TransCatLimit{
			Category: str,
			Active:   true,
			UserId:   BotUsers[q.From.UserName].UserId,
		}
		switch split[2] {
		case "debit":
			cat.Direction = -1
		case "deposit":
			cat.Direction = 0
		case "credit":
			cat.Direction = 1
		}
		b.accountant.EditCats(ctx, cat)
		requestCategoriesKeyboardEditor(b, ctx, 0, &userChat{q.Message.Chat.ID, q.Message.MessageID, q.From.UserName})
	case "limit":
		b.requestReply(q, "REC_NEWLIMIT")
	case "disable":
		cat := &entity.TransCatLimit{
			Category: q.Message.Text,
			Active:   false,
			UserId:   BotUsers[q.From.UserName].UserId,
		}
		b.accountant.EditCats(ctx, cat)
		requestCategoriesKeyboardEditor(b, ctx, 0, &userChat{q.Message.Chat.ID, q.Message.MessageID, q.From.UserName})
	default:
		requestCategoryKeyboardEditor(b, ctx, q)
	}
}

func getDebitCreditKeyboard() *tgbotapi.InlineKeyboardMarkup {
	mrkp := newKeyboardForm()
	mrkp.setOptions([][]string{
		{"Доходы " + EMOJI_CREDIT, PREFIX_SETCATEGORY + ":new:credit"},
		{"Сбережения " + EMOJI_DEPOSIT, PREFIX_SETCATEGORY + ":new:deposit"},
		{"Расходы " + EMOJI_DEBIT, PREFIX_SETCATEGORY + ":new:debit"},
	})
	res, err := mrkp.getMarkup()
	if err != nil {
		return nil
	}
	return res
}

func requestCategoryKeyboardEditor(b *Bot, ctx context.Context, q *tgbotapi.CallbackQuery) {
	category := strings.Split(q.Data, ":")
	if len(category) < 2 {
		return
	}
	b.updateMsgText(q.Message.Chat.ID, q.Message.MessageID, category[1])

	mrkp := newKeyboardForm()
	mrkp.setOptions([][]string{
		{"Задать лимит", PREFIX_SETCATEGORY + ":limit:" + category[1]},
		{"Сделать неактивной", PREFIX_SETCATEGORY + ":disable:" + category[1]},
	})
	mrkp.setControl([][][]string{
		{{EMOJI_HOOK_BACK, PREFIX_SETCATEGORY + ":backToCategories"}},
	})
	resMrkp, err := mrkp.getMarkup()
	if err != nil {
		fmt.Println(err)
		return
	}
	msg := tgbotapi.NewEditMessageReplyMarkup(q.Message.Chat.ID, q.Message.MessageID, *resMrkp)
	b.api.Send(msg)
}

func requestCategoriesKeyboardEditor(b *Bot, ctx context.Context, page int, c *userChat) {
	b.updateMsgText(c.chatID, c.messageID, "Настройки категорий и лимитов")
	cats, err := b.accountant.GetCatsLimit(ctx, BotUsers[c.userName].UserId)
	if err != nil {
		return
	}

	options := make([][]string, len(cats), len(cats)+1)
	var catDirection, catLimit string
	for i, v := range cats {
		switch v.Direction {
		case -1:
			catDirection = EMOJI_DEBIT
		case 0:
			catDirection = EMOJI_DEPOSIT
		case 1:
			catDirection = EMOJI_CREDIT
		}
		if v.Limit > 0 {
			catLimit = fmt.Sprintf(" (%d₽)", v.Limit)
		} else {
			catLimit = ""
		}
		options[i] = []string{
			v.Category + " " + catDirection + catLimit,
			PREFIX_SETCATEGORY + ":" + v.Category,
		}
	}
	options = append(options, []string{EMOJI_PLUS, PREFIX_SETCATEGORY + ":addNewCat"})
	mrkp := newKeyboardForm()
	mrkp.setOptions(options)
	mrkp.addNavigationControl(page, []string{EMOJI_HOOK_BACK, PREFIX_SETCATEGORY + ":backToSettings"}, nil)
	resMrkp, err := mrkp.getMarkup()
	if err != nil {
		fmt.Println(err)
		return
	}

	msg := tgbotapi.NewEditMessageReplyMarkup(c.chatID, c.messageID, *resMrkp)
	b.api.Send(msg)
}
