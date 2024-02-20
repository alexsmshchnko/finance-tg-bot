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

func (b *Bot) callbackQueryHandler(ctx context.Context, query *tgbotapi.CallbackQuery) {
	split := strings.Split(query.Data, ":")
	switch split[0] {
	case PREFIX_CATEGORY:
		b.handleCategoryCallbackQuery(query)
	case PREFIX_SUBCATEGORY:
		b.handleSubCategoryCallbackQuery(query)
	case PREFIX_OPTION:
		b.handleOptionCallbackQuery(query)
	case PREFIX_PAGE:
		b.handleNavigationCallbackQuery(ctx, query)
	case PREFIX_REPORT:
		b.handleReportCallbackQuery(ctx, query)
	case PREFIX_SETTING:
		b.handleSettingCallbackQuery(ctx, query)
	case PREFIX_SETCATEGORY:
		b.handleCategoryKeyboardEditor(ctx, query)
	}
}

func (b *Bot) handleCategoryCallbackQuery(query *tgbotapi.CallbackQuery) {
	cat, _ := strings.CutPrefix(query.Data, PREFIX_CATEGORY+":")

	query.Message.Text, _ = strings.CutSuffix(query.Message.Text, "₽")

	//update text
	b.updateMsgText(query.Message.Chat.ID, query.Message.MessageID, query.Message.Text+"₽ на "+cat)

	//query description
	b.requestSubCats(context.Background(), 0, query)
}

func (b *Bot) handleSubCategoryCallbackQuery(query *tgbotapi.CallbackQuery) {
	//update text
	subCat, _ := strings.CutPrefix(query.Data, PREFIX_SUBCATEGORY+":")

	if subCat == "writeCustom" {
		b.requestReply(query, "REC_DESC")
		return
	}

	b.updateMsgText(query.Message.Chat.ID, query.Message.MessageID, query.Message.Text+"\n"+EMOJI_COMMENT+subCat)

	//update keyboard
	mrkp := getMsgOptionsKeyboard()
	msg := tgbotapi.NewEditMessageReplyMarkup(query.Message.Chat.ID, query.Message.MessageID, *mrkp)
	b.api.Send(msg)
}

func (b *Bot) handleOptionCallbackQuery(query *tgbotapi.CallbackQuery) {
	split := strings.Split(query.Data, ":")
	switch split[1] {
	case "saveRecord":
		b.confirmRecord(query)
	// case "addDescription":
	// 	b.requestDescription(query)
	case "deleteRecord":
		b.deleteRecord(query)
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
		b.updateMsgText(u.Message.Chat.ID, respMsg.MessageID, u.Message.Text)
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
		err = b.accountant.EditCats(cat, u.SentFrom().UserName)
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
