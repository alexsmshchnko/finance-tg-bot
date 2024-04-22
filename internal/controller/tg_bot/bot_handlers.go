package tg_bot

import (
	"context"
	"database/sql"
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
		b.handleCategoryCallbackQuery(q)
	case PREFIX_SUBCATEGORY:
		b.handleSubCategoryCallbackQuery(q)
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

func (b *Bot) handleCategoryCallbackQuery(q *tgbotapi.CallbackQuery) {
	cat, _ := strings.CutPrefix(q.Data, PREFIX_CATEGORY+":")

	q.Message.Text, _ = strings.CutSuffix(q.Message.Text, "₽")

	//update text
	b.updateMsgText(q.Message.Chat.ID, q.Message.MessageID, q.Message.Text+"₽ на "+cat)

	//query description
	b.requestSubCats(context.Background(), 0, q)
}

func (b *Bot) handleSubCategoryCallbackQuery(q *tgbotapi.CallbackQuery) {
	//update text
	subCat, _ := strings.CutPrefix(q.Data, PREFIX_SUBCATEGORY+":")

	if subCat == "writeCustom" {
		b.requestReply(q, "REC_DESC")
		return
	}

	b.updateMsgText(q.Message.Chat.ID, q.Message.MessageID, q.Message.Text+"\n"+EMOJI_COMMENT+subCat)

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
			return
		}
		stat, err := b.accountant.Money2Time(amnt, BotUsers[q.From.UserName].UserId)
		if err != nil {
			return
		}
		msg := tgbotapi.NewMessage(q.Message.Chat.ID, fmt.Sprintf("%s\n\n%s", q.Message.Text, stat))
		b.api.Send(msg)
	}
}

func (b *Bot) responseHandler(u *tgbotapi.Update) {
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
			fmt.Println(err)
			return
		}
		cat := entity.TransCatLimit{
			Category:  sql.NullString{String: respMsg.Text, Valid: true},
			Direction: sql.NullInt16{Int16: 0, Valid: false},
			Active:    sql.NullBool{Bool: true, Valid: true},
			Limit:     sql.NullInt64{Int64: int64(limit), Valid: true},
		}
		err = b.accountant.EditCats(context.Background(), cat, u.SentFrom().UserName)
		if err != nil {
			fmt.Println(err)
			return
		}
		requestCategoriesKeyboardEditor(b, context.Background(), 0, &userChat{u.Message.Chat.ID, respMsg.MessageID, u.SentFrom().UserName})
	}

	waitUserResponseComplete(u.SentFrom().UserName)
	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID)
	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID-1)
}
