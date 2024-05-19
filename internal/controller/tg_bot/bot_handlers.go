package tg_bot

import (
	"context"
	"finance-tg-bot/internal/entity"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) callbackQueryHandler(ctx context.Context, q *tgbotapi.CallbackQuery) {
	split := strings.Split(q.Data, ":")
	switch split[0] {
	case PREFIX_CATEGORY:
		b.handleCategoryCallbackQuery(ctx, q)
	case PREFIX_SUBCATEGORY:
		b.handleSubCategoryCallbackQuery(ctx, q)
	case PREFIX_OPTION:
		b.handleOptionCallbackQuery(q)
	case PREFIX_PAGE:
		b.handleNavigationCallbackQuery(ctx, q)
	case PREFIX_REPORT:
		b.handleReportCallbackQuery(q)
	case PREFIX_SETTING:
		b.handleSettingCallbackQuery(ctx, q)
	case PREFIX_SETCATEGORY:
		b.handleCategoryKeyboardEditor(ctx, q)
	}
}

func (b *Bot) handleCategoryCallbackQuery(ctx context.Context, q *tgbotapi.CallbackQuery) {
	splt := strings.Split(q.Data, ":")
	if len(splt) < 2 {
		b.log.Error("handleCategoryCallbackQuery", "cat Data short request (<2)", q.Data)
		return
	}

	if splt[1] == "cancel" {
		b.deleteMsg(q.Message.Chat.ID, q.Message.MessageID)
		return
	}

	//update finMsg
	finMsg, err := NewFinMsg().parseFinMsg(q.Message.Text)
	if err != nil {
		b.log.Error("handleCategoryCallbackQuery parseFinMsg", "err", err)
		return
	}
	finMsg.SetCategory(splt[1])

	b.updateMsgText(q.Message.Chat.ID, q.Message.MessageID, finMsg.String())

	//query description
	b.requestSubCats(ctx, 0, q)
}

func (b *Bot) handleSubCategoryCallbackQuery(ctx context.Context, q *tgbotapi.CallbackQuery) {
	splt := strings.Split(q.Data, ":")
	if len(splt) < 2 {
		b.log.Error("handleSubCategoryCallbackQuery", "subCat Data short request (<2)", q.Data)
		return
	}

	switch splt[1] {
	case "writeCustom":
		b.requestReply(q, "REC_DESC")
		return
	case "backToCategories":
		finMsg, err := NewFinMsg().parseFinMsg(q.Message.Text)
		if err != nil {
			b.log.Error("handleSubCategoryCallbackQuery parseFinMsg", "err", err)
			return
		}
		finMsg.SetCategory("")
		b.updateMsgText(q.Message.Chat.ID, q.Message.MessageID, finMsg.String())

		b.requestCats(ctx, 0,
			&userChat{q.Message.Chat.ID, q.Message.MessageID, q.From.UserName, ""})
		return
	}

	if len(splt) < 3 {
		b.log.Error("handleSubCategoryCallbackQuery", "subCat Data short request (<3)", q.Data)
		return
	}
	idx, err := strconv.Atoi(splt[2])
	if err != nil {
		b.log.Error("handleSubCategoryCallbackQuery idx strconv.Atoi", "err", err)
		return
	}
	subCats, err := b.accountant.GetSubCats(ctx, BotUsers[q.From.UserName].UserId, splt[1])
	if err != nil {
		b.log.Error("handleSubCategoryCallbackQuery b.accountant.GetSubCats", "err", err)
		return
	} else if len(subCats) < idx+1 {
		b.log.Error("handleSubCategoryCallbackQuery short array subCats")
		return
	}

	finMsg, err := NewFinMsg().parseFinMsg(q.Message.Text)
	if err != nil {
		b.log.Error("handleSubCategoryCallbackQuery parseFinMsg", "err", err)
		return
	}
	finMsg.SetDescription(subCats[idx])

	b.updateMsgText(q.Message.Chat.ID, q.Message.MessageID, finMsg.String())

	//update keyboard
	b.api.Send(tgbotapi.NewEditMessageReplyMarkup(q.Message.Chat.ID, q.Message.MessageID, getMsgOptionsKeyboard()))
}

func (b *Bot) handleOptionCallbackQuery(q *tgbotapi.CallbackQuery) {
	split := strings.Split(q.Data, ":")
	switch split[1] {
	case "saveRecord":
		b.confirmRecord(q)
	case "expandOptions":
		b.api.Send(tgbotapi.NewEditMessageReplyMarkup(q.Message.Chat.ID, q.Message.MessageID, getMsgExpOptionsKeyboard()))
	case "deleteRecord":
		b.deleteRecord(q)
	case "money2Time":
		finMsg, err := NewFinMsg().parseFinMsg(q.Message.Text)
		if err != nil {
			b.log.Error("handleOptionCallbackQuery parseFinMsg", "err", err)
			return
		}
		stat, err := b.accountant.Money2Time(finMsg.amount, BotUsers[q.From.UserName].UserId)
		if err != nil {
			b.log.Error("handleOptionCallbackQuery Money2Time", "err", err)
			return
		}
		b.api.Send(tgbotapi.NewMessage(q.Message.Chat.ID, fmt.Sprintf("%s\n\n%s", q.Message.Text, stat)))
	}
}

func (b *Bot) handleNavigationCallbackQuery(ctx context.Context, q *tgbotapi.CallbackQuery) {
	split := strings.Split(q.Data, ":")
	page, err := strconv.Atoi(split[2])
	if err != nil {
		b.log.Error("handleNavigationCallbackQuery page strconv.Atoi", "err", err)
		return
	}

	switch split[1] {
	case "next":
		page++
	case "prev":
		page--
	}

	switch strings.Split(*q.Message.ReplyMarkup.InlineKeyboard[0][0].CallbackData, ":")[0] {
	case PREFIX_CATEGORY:
		b.requestCats(ctx, page,
			&userChat{q.Message.Chat.ID, q.Message.MessageID, q.From.UserName, ""})
	case PREFIX_SUBCATEGORY:
		b.requestSubCats(ctx, page, q)
	case PREFIX_SETCATEGORY:
		requestCategoriesKeyboardEditor(b, ctx, page,
			&userChat{q.Message.Chat.ID, q.Message.MessageID, q.From.UserName, ""})
	}
}

func (b *Bot) responseHandler(ctx context.Context, u *tgbotapi.Update) {
	respMsg := BotUsers[u.SentFrom().UserName].ResponseMsg

	switch BotUsers[u.SentFrom().UserName].ResponseCode {
	case "REC_DESC":
		finMsg, err := NewFinMsg().parseFinMsg(respMsg.Text)
		if err != nil {
			b.log.Error("handleOptionCallbackQuery parseFinMsg", "err", err)
			return
		}
		finMsg.SetDescription(u.Message.Text)
		b.updateMsgText(u.Message.Chat.ID, respMsg.MessageID, finMsg.String())
		b.api.Send(tgbotapi.NewEditMessageReplyMarkup(u.Message.Chat.ID, respMsg.MessageID, getMsgOptionsKeyboard()))
	case "REC_NEWCAT":
		b.updateMsgText(u.Message.Chat.ID, respMsg.MessageID, "Тип траты для "+u.Message.Text)
		b.api.Send(tgbotapi.NewEditMessageReplyMarkup(u.Message.Chat.ID, respMsg.MessageID, getDebitCreditKeyboard()))
	case "REC_NEWLIMIT":
		limit, err := strconv.Atoi(u.Message.Text)
		if err != nil {
			b.log.Error("responseHandler limit strconv.Atoi", "err", err)
			return
		}
		cat := &entity.TransCatLimit{
			UserId:   BotUsers[u.SentFrom().UserName].UserId,
			Category: respMsg.Text,
			Active:   true,
			Limit:    limit,
		}
		err = b.accountant.EditCats(ctx, cat)
		if err != nil {
			b.log.Error("responseHandler EditCats", "err", err)
			return
		}
		requestCategoriesKeyboardEditor(b, ctx, 0,
			&userChat{u.Message.Chat.ID, respMsg.MessageID, u.SentFrom().UserName, ""})
	}

	waitUserResponseComplete(b.log, u.SentFrom().UserName)
	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID)
	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID-1)
}

func (b *Bot) replyHandler(u *tgbotapi.Update) {
	if amnt, err := strconv.Atoi(u.Message.Text); err == nil {
		// message amount update
		finMsg, err := NewFinMsg().parseFinMsg(u.Message.ReplyToMessage.Text)
		if err != nil {
			b.log.Error("replyHandler parseFinMsg", "err", err)
			return
		}
		b.api.Send(tgbotapi.NewMessage(u.Message.Chat.ID, fmt.Sprintf("%d -> %d", finMsg.amount, amnt)))

		finMsg.SetAmount(amnt)

		b.api.Send(tgbotapi.NewEditMessageText(u.Message.Chat.ID, u.Message.ReplyToMessage.MessageID, finMsg.String()))
	}
	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID)
	b.api.Send(tgbotapi.NewEditMessageReplyMarkup(u.Message.Chat.ID, u.Message.ReplyToMessage.MessageID, getMsgOptionsKeyboard()))
}
