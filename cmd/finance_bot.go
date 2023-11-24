package main

import (
	"context"
	"fmt"
	"log"

	// "log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"finance-tg-bot/internal/accountant"
	"finance-tg-bot/internal/config"
	"finance-tg-bot/internal/disk"
	"finance-tg-bot/internal/storage"
	"finance-tg-bot/internal/synchronizer"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	gBot  *tgbotapi.BotAPI
	db    *storage.PGStorage
	cloud *disk.Disk
	acnt  *accountant.Accountant
	sync  *synchronizer.Synchronizer

	// log    *slog.Logger
	ctx    context.Context
	cancel context.CancelFunc
)

func deleteMsg(chatID int64, messageID int) {
	gBot.Send(tgbotapi.NewDeleteMessage(chatID, messageID))
}

func clearMsgReplyMarkup(chatID int64, messageID int) {
	mrkp := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0),
	}

	msg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, mrkp)
	gBot.Send(msg)
}

func updateMsgText(chatID int64, messageID int, text string) {
	gBot.Send(tgbotapi.NewEditMessageText(chatID, messageID, text))
}

func processNumber(u *tgbotapi.Update) (err error) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, u.Message.Text+"₽")
	cats := BotUsers[u.Message.Chat.UserName].FinCategories
	if len(cats) < 1 {
		cats, err = acnt.GetCats(ctx, u.Message.Chat.UserName)
		if err != nil {
			return err
		}
		BotUsers[u.Message.Chat.UserName] = BotUser{FinCategories: cats}

	}
	msg.ReplyMarkup = getCatPageInlineKeyboard(cats, 0)
	_, err = gBot.Send(msg)
	deleteMsg(u.Message.Chat.ID, u.Message.MessageID)

	return err
}

func confirmRecord(query *tgbotapi.CallbackQuery) {
	amnt, _ := strconv.Atoi(strings.Split(query.Message.Text, "₽")[0])
	cat := strings.Split(query.Message.Text, " на ")[1]

	acnt.PostDoc(cat, amnt, "", fmt.Sprint(query.Message.MessageID), query.From.UserName)
	clearMsgReplyMarkup(query.Message.Chat.ID, query.Message.MessageID)
}

func deleteRecord(query *tgbotapi.CallbackQuery) {
	acnt.DeleteDoc(fmt.Sprint(query.Message.MessageID), query.From.UserName)
	deleteMsg(query.Message.Chat.ID, query.Message.MessageID)
}

// func addDescription(u *tgbotapi.Update) (err error) {
// 	rec := internal.ReceiptRec{Description: u.Message.Text}
// 	err = internal.AddLastExpenseDescription(&rec)
// 	if err != nil {
//

// 	updateMsgText(u.Message.Chat.ID, u.Message.ReplyToMessage.MessageID, u.Message.ReplyToMessage.Text+"\n"+EMOJI_COMMENT+u.Message.Text)

// 	deleteMsg(u.Message.Chat.ID, u.Message.MessageID)

//

func requestDescription(query *tgbotapi.CallbackQuery) {
	// msg := tgbotapi.NewMessage(u.CallbackQuery.Message.Chat.ID, "описание")
	// msg.ReplyMarkup = getReply()

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, EMOJI_COMMENT+"...")

	msg.ReplyMarkup = getReply()
	gBot.Send(msg)
	WaitUserResponeStart(query.From.UserName, "REC_DESC", *query.Message)
}

func handleCategoryCallbackQuery(query *tgbotapi.CallbackQuery) {
	// cat := strings.Split(query.Data, "CAT:")[1]
	cat, _ := strings.CutPrefix(query.Data, "CAT:")
	query.Message.Text, _ = strings.CutSuffix(query.Message.Text, "₽")

	resp := query.Message.Text + "₽ на " + cat
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, resp)
	msg.ReplyMarkup = getMsgOptionsKeyboard()
	gBot.Send(msg)
	deleteMsg(query.Message.Chat.ID, query.Message.MessageID)

	// rec := internal.NewReceiptRec(time.Now(), cat, amnt, "")
	// internal.AddNewExpense(rec)
}

func handleNavigationCallbackQuery(query *tgbotapi.CallbackQuery) {
	var err error
	cats := BotUsers[query.From.UserName].FinCategories
	if len(cats) < 1 {
		fmt.Println("User category cash is empty")
		cats, err = acnt.GetCats(ctx, query.From.UserName)
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

	gBot.Send(msg)

}

func handleOptionCallbackQuery(query *tgbotapi.CallbackQuery) {
	switch query.Data {
	case "OPT:saveRecord":
		confirmRecord(query)
	case "OPT:addDescription":
		requestDescription(query)
	case "OPT:deleteRecord":
		deleteRecord(query)
	}
}

func callbackQueryHandler(query *tgbotapi.CallbackQuery) {
	split := strings.Split(query.Data, ":")
	switch split[0] {
	case "CAT":
		handleCategoryCallbackQuery(query)
	case "OPT":
		handleOptionCallbackQuery(query)
	case "PAGE":
		handleNavigationCallbackQuery(query)
	}
}

// func processReply(u *tgbotapi.Update) (err error) {
// 	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "")
// 	msg.ReplyMarkup = getReply()

// 	gBot.Send(msg)

// 	return
// }

func processResponse(u *tgbotapi.Update) {
	//msg := tgbotapi.NewMessage(u.Message.Chat.ID, u.Message.Text+"response processed")

	// rec := internal.ReceiptRec{Description: u.Message.Text}
	// err := internal.AddLastExpenseDescription(&rec)
	// if err != nil {
	//

	respMsg := BotUsers[u.SentFrom().UserName].ResponseMsg
	updateMsgText(u.Message.Chat.ID, respMsg.MessageID, respMsg.Text+"\n"+EMOJI_COMMENT+u.Message.Text)

	amnt, _ := strconv.Atoi(strings.Split(respMsg.Text, "₽")[0])
	cat := strings.Split(respMsg.Text, " на ")[1]

	acnt.PostDoc(cat, amnt, u.Message.Text, fmt.Sprint(respMsg.MessageID), u.SentFrom().UserName)

	// expRec := internal.NewFinRec(cat, amnt, u.Message.Text, fmt.Sprintf("%d", respMsg.MessageID))
	// internal.NewUser(u.SentFrom().UserName).NewExpense(expRec)

	WaitUserResponseComplete(u.SentFrom().UserName)

	deleteMsg(u.Message.Chat.ID, u.Message.MessageID)
	deleteMsg(u.Message.Chat.ID, u.Message.MessageID-1)
	//gBot.Send(msg)
}

func checkUser(userName string) bool {
	_, f := BotUsers[userName]
	if !f {
		active, err := acnt.GetUserStatus(ctx, userName)
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

func handleUpdate(update *tgbotapi.Update) {

	if !checkUser(update.SentFrom().UserName) {
		return
	}

	if update.CallbackQuery != nil {
		callbackQueryHandler(update.CallbackQuery)
		return
	}

	if update.Message != nil {
		if ResponseIsAwaited(update.SentFrom().UserName) {
			processResponse(update)
		}

		// if update.Message.ReplyToMessage != nil {
		// 	processReply(&update)
		// }

		if update.Message.IsCommand() {
			processCommand(update)
		}

		if len(update.Message.Text) > 0 {
			if _, err := strconv.Atoi(update.Message.Text); err == nil {
				processNumber(update)
			}
		}
	}
}

func runBot(ctx context.Context) (err error) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := gBot.GetUpdatesChan(updateConfig)

	gBot.Send(initCommands())

	for {
		select {
		case update := <-updates:
			handleUpdate(&update)
		case <-ctx.Done():
			gBot.StopReceivingUpdates()
			return ctx.Err()
		}
	}
}

func run() (err error) {
	if gBot, err = tgbotapi.NewBotAPI(config.Get().TelegramBotToken); err != nil {
		log.Println("[ERROR] failed to create botAPI")
		return
	}
	gBot.Debug = true

	log.Printf("Authorized on account %s", gBot.Self.UserName)

	ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db = storage.NewPGStorage(ctx, config.Get().DatabaseDSN)
	cloud = disk.New()

	acnt = accountant.NewAccountant(db)
	sync = synchronizer.New(cloud, db)

	err = runBot(ctx)

	return
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
