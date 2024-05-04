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
		b.handleReportCallbackQuery(ctx, q)
	case PREFIX_SETTING:
		b.handleSettingCallbackQuery(ctx, q)
	case PREFIX_SETCATEGORY:
		b.handleCategoryKeyboardEditor(ctx, q)
	}
}

func (b *Bot) handleCategoryCallbackQuery(ctx context.Context, q *tgbotapi.CallbackQuery) {
	cat, _ := strings.CutPrefix(q.Data, PREFIX_CATEGORY+":")

	q.Message.Text, _ = strings.CutSuffix(q.Message.Text, "₽")

	//update text
	b.updateMsgText(q.Message.Chat.ID, q.Message.MessageID, q.Message.Text+"₽ на "+cat)

	//query description
	b.requestSubCats(ctx, 0, q)
}

func (b *Bot) handleSubCategoryCallbackQuery(ctx context.Context, q *tgbotapi.CallbackQuery) {
	splt := strings.Split(q.Data, ":")

	if len(splt) < 2 {
		b.log.Error("handleSubCategoryCallbackQuery", "subCat Data short request (<2)", q.Data)
		return
	}

	if splt[1] == "writeCustom" {
		b.requestReply(q, "REC_DESC")
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
	b.updateMsgText(q.Message.Chat.ID, q.Message.MessageID, q.Message.Text+"\n"+EMOJI_COMMENT+subCats[idx])

	//update keyboard
	mrkp := getMsgOptionsKeyboard()
	msg := tgbotapi.NewEditMessageReplyMarkup(q.Message.Chat.ID, q.Message.MessageID, *mrkp)
	b.api.Send(msg)
}

func (b *Bot) handleOptionCallbackQuery(q *tgbotapi.CallbackQuery) {
	split := strings.Split(q.Data, ":")
	switch split[1] {
	case "saveRecord":
		b.confirmRecord(q)
	case "expandOptions":
		msg := tgbotapi.NewEditMessageReplyMarkup(q.Message.Chat.ID, q.Message.MessageID, *getMsgExpOptionsKeyboard())
		b.api.Send(msg)
	case "deleteRecord":
		b.deleteRecord(q)
	case "money2Time":
		amnt, err := strconv.Atoi(strings.Split(q.Message.Text, "₽")[0])
		if err != nil {
			b.log.Error("handleOptionCallbackQuery amnt strconv.Atoi", "err", err)
			return
		}
		stat, err := b.accountant.Money2Time(amnt, BotUsers[q.From.UserName].UserId)
		if err != nil {
			b.log.Error("handleOptionCallbackQuery Money2Time", "err", err)
			return
		}
		msg := tgbotapi.NewMessage(q.Message.Chat.ID, fmt.Sprintf("%s\n\n%s", q.Message.Text, stat))
		b.api.Send(msg)
	}
}

func (b *Bot) responseHandler(ctx context.Context, u *tgbotapi.Update) {
	respCode := BotUsers[u.SentFrom().UserName].ResponseCode
	respMsg := BotUsers[u.SentFrom().UserName].ResponseMsg
	switch respCode {
	case "REC_DESC":
		b.updateMsgText(u.Message.Chat.ID, respMsg.MessageID, respMsg.Text+"\n"+EMOJI_COMMENT+u.Message.Text)
		mrkp := getMsgOptionsKeyboard()
		msg := tgbotapi.NewEditMessageReplyMarkup(u.Message.Chat.ID, respMsg.MessageID, *mrkp)
		b.api.Send(msg)
	case "REC_NEWCAT":
		b.updateMsgText(u.Message.Chat.ID, respMsg.MessageID, "Тип траты для "+u.Message.Text)
		mrkp := getDebitCreditKeyboard()
		msg := tgbotapi.NewEditMessageReplyMarkup(u.Message.Chat.ID, respMsg.MessageID, *mrkp)
		b.api.Send(msg)
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
		requestCategoriesKeyboardEditor(b, ctx, 0, &userChat{u.Message.Chat.ID, respMsg.MessageID, u.SentFrom().UserName})
	}

	waitUserResponseComplete(u.SentFrom().UserName)
	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID)
	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID-1)
}
