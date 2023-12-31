package tg_bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type accountant interface {
	GetCats(ctx context.Context, username string) (cats []string, err error)
	GetUserStatus(ctx context.Context, username string) (status bool, err error)
	PostDoc(category string, amount int, description string, msg_id string, client string) (err error)
	DeleteDoc(msg_id string, client string) (err error)
}

type synchronizer interface {
	MigrateFromCloud(ctx context.Context, username string) (err error)
}

type Bot struct {
	api *tgbotapi.BotAPI
	accountant
	synchronizer
}

func New(api *tgbotapi.BotAPI, acc accountant, sync synchronizer) *Bot {
	return &Bot{
		api:          api,
		accountant:   acc,
		synchronizer: sync,
	}
}

func (b *Bot) deleteMsg(chatID int64, messageID int) {
	b.api.Send(tgbotapi.NewDeleteMessage(chatID, messageID))
}

func (b *Bot) clearMsgReplyMarkup(chatID int64, messageID int) {
	mrkp := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0),
	}

	msg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, mrkp)
	b.api.Send(msg)
}

func (b *Bot) updateMsgText(chatID int64, messageID int, text string) {
	b.api.Send(tgbotapi.NewEditMessageText(chatID, messageID, text))
}

func (b *Bot) processNumber(ctx context.Context, u *tgbotapi.Update) (err error) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, u.Message.Text+"₽")
	cats := BotUsers[u.Message.Chat.UserName].FinCategories
	if len(cats) < 1 {
		cats, err = b.accountant.GetCats(ctx, u.Message.Chat.UserName)
		if err != nil {
			return err
		}
		BotUsers[u.Message.Chat.UserName] = BotUser{FinCategories: cats}

	}
	msg.ReplyMarkup = getCatPageInlineKeyboard(cats, 0)
	_, err = b.api.Send(msg)
	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID)

	return err
}

func (b *Bot) confirmRecord(query *tgbotapi.CallbackQuery) {
	amnt, _ := strconv.Atoi(strings.Split(query.Message.Text, "₽")[0])
	cat := strings.Split(query.Message.Text, " на ")[1]

	b.accountant.PostDoc(cat, amnt, "", fmt.Sprint(query.Message.MessageID), query.From.UserName)
	b.clearMsgReplyMarkup(query.Message.Chat.ID, query.Message.MessageID)
}

func (b *Bot) deleteRecord(query *tgbotapi.CallbackQuery) {
	b.accountant.DeleteDoc(fmt.Sprint(query.Message.MessageID), query.From.UserName)
	b.deleteMsg(query.Message.Chat.ID, query.Message.MessageID)
}

// func addDescription(u *tgbotapi.Update) (err error) {
// 	rec := internal.ReceiptRec{Description: u.Message.Text}
// 	err = internal.AddLastExpenseDescription(&rec)
// 	if err != nil {
//

// 	updateMsgText(u.Message.Chat.ID, u.Message.ReplyToMessage.MessageID, u.Message.ReplyToMessage.Text+"\n"+EMOJI_COMMENT+u.Message.Text)

// 	deleteMsg(u.Message.Chat.ID, u.Message.MessageID)

//

func (b *Bot) requestDescription(query *tgbotapi.CallbackQuery) {
	// msg := tgbotapi.NewMessage(u.CallbackQuery.Message.Chat.ID, "описание")
	// msg.ReplyMarkup = getReply()

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, EMOJI_COMMENT+"...")

	msg.ReplyMarkup = getReply()
	b.api.Send(msg)
	WaitUserResponeStart(query.From.UserName, "REC_DESC", *query.Message)
}

func (b *Bot) handleCategoryCallbackQuery(query *tgbotapi.CallbackQuery) {
	// cat := strings.Split(query.Data, "CAT:")[1]
	cat, _ := strings.CutPrefix(query.Data, "CAT:")
	query.Message.Text, _ = strings.CutSuffix(query.Message.Text, "₽")

	resp := query.Message.Text + "₽ на " + cat
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, resp)
	msg.ReplyMarkup = getMsgOptionsKeyboard()
	b.api.Send(msg)
	b.deleteMsg(query.Message.Chat.ID, query.Message.MessageID)

	// rec := internal.NewReceiptRec(time.Now(), cat, amnt, "")
	// internal.AddNewExpense(rec)
}

func (b *Bot) handleNavigationCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) {
	var err error
	cats := BotUsers[query.From.UserName].FinCategories
	if len(cats) < 1 {
		fmt.Println("User category cash is empty")
		cats, err = b.accountant.GetCats(ctx, query.From.UserName)
		if err != nil {
			log.Println(err)
		}
	}

	split := strings.Split(query.Data, ":")
	page, err := strconv.Atoi(split[2])
	if err != nil {
		log.Println(err)
	}
	switch split[1] {
	case "next":
		page++
	case "prev":
		page--
	}

	mrkp := getCatPageInlineKeyboard(cats, page)

	msg := tgbotapi.NewEditMessageReplyMarkup(query.Message.Chat.ID, query.Message.MessageID, *mrkp)

	// msg.ReplyMarkup =

	// mrkp := tgbotapi.InlineKeyboardMarkup{
	// 	InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0),
	// }

	b.api.Send(msg)

}

func (b *Bot) handleOptionCallbackQuery(query *tgbotapi.CallbackQuery) {
	switch query.Data {
	case "OPT:saveRecord":
		b.confirmRecord(query)
	case "OPT:addDescription":
		b.requestDescription(query)
	case "OPT:deleteRecord":
		b.deleteRecord(query)
	}
}

func (b *Bot) callbackQueryHandler(ctx context.Context, query *tgbotapi.CallbackQuery) {
	split := strings.Split(query.Data, ":")
	switch split[0] {
	case "CAT":
		b.handleCategoryCallbackQuery(query)
	case "OPT":
		b.handleOptionCallbackQuery(query)
	case "PAGE":
		b.handleNavigationCallbackQuery(ctx, query)
	}
}

// func processReply(u *tgbotapi.Update) (err error) {
// 	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "")
// 	msg.ReplyMarkup = getReply()

// 	gBot.Send(msg)

// 	return
// }

func (b *Bot) processResponse(u *tgbotapi.Update) {
	//msg := tgbotapi.NewMessage(u.Message.Chat.ID, u.Message.Text+"response processed")

	// rec := internal.ReceiptRec{Description: u.Message.Text}
	// err := internal.AddLastExpenseDescription(&rec)
	// if err != nil {
	//

	respMsg := BotUsers[u.SentFrom().UserName].ResponseMsg
	b.updateMsgText(u.Message.Chat.ID, respMsg.MessageID, respMsg.Text+"\n"+EMOJI_COMMENT+u.Message.Text)

	amnt, _ := strconv.Atoi(strings.Split(respMsg.Text, "₽")[0])
	cat := strings.Split(respMsg.Text, " на ")[1]

	b.accountant.PostDoc(cat, amnt, u.Message.Text, fmt.Sprint(respMsg.MessageID), u.SentFrom().UserName)

	// expRec := internal.NewFinRec(cat, amnt, u.Message.Text, fmt.Sprintf("%d", respMsg.MessageID))
	// internal.NewUser(u.SentFrom().UserName).NewExpense(expRec)

	WaitUserResponseComplete(u.SentFrom().UserName)

	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID)
	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID-1)
	//gBot.Send(msg)
}

func (b *Bot) checkUser(ctx context.Context, userName string) bool {
	_, f := BotUsers[userName]
	if !f {
		active, err := b.accountant.GetUserStatus(ctx, userName)
		if err != nil {
			log.Println(err)
		}
		if active {
			NewBotUser(userName)
		}
		return active
	}
	return f
}

func (b *Bot) handleUpdate(ctx context.Context, update *tgbotapi.Update) {

	if !b.checkUser(ctx, update.SentFrom().UserName) {
		return
	}

	if update.CallbackQuery != nil {
		b.callbackQueryHandler(ctx, update.CallbackQuery)
		return
	}

	if update.Message != nil {
		if ResponseIsAwaited(update.SentFrom().UserName) {
			b.processResponse(update)
		}

		// if update.Message.ReplyToMessage != nil {
		// 	processReply(&update)
		// }

		if update.Message.IsCommand() {
			b.processCommand(ctx, update)
		}

		if len(update.Message.Text) > 0 {
			if _, err := strconv.Atoi(update.Message.Text); err == nil {
				b.processNumber(ctx, update)
			}
		}
	}
}

func (b *Bot) Run(ctx context.Context) (err error) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := b.api.GetUpdatesChan(updateConfig)

	b.api.Send(initCommands())

	for {
		select {
		case update := <-updates:
			ctxU, cancelU := context.WithTimeout(ctx, 3*time.Second)
			b.handleUpdate(ctxU, &update)
			cancelU()
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return ctx.Err()
		}
	}
}
