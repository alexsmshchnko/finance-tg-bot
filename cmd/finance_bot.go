package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"finance-tg-bot/internal/accountant"
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
		cats, err = acnt.GetCats(context.Background(), u.Message.Chat.UserName)
		if err != nil {
			return err
		}
	}
	msg.ReplyMarkup = getCatInlineKeyboard(cats, 0, 1)
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
	// cats := BotUsers[query.Message.From.UserName].FinCategories
	// if len(cats) < 1 {
	cats, err := acnt.GetCats(context.Background(), query.From.UserName)
	if err != nil {
		log.Println(err)
	}
	// }

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

	mrkp := getCatInlineKeyboard(cats, page, 1)

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

	// msg := tgbotapi.NewMessage(query.Message.Chat.ID, "")
	// gBot.Send(msg)
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
	return f
}

func run() (err error) {
	//CONFIG
	if gBot, err = tgbotapi.NewBotAPI(os.Getenv(BOT_TOKEN_NAME)); err != nil {
		log.Println("[ERROR] failed to create botAPI")
		return
	}
	gBot.Debug = true

	log.Printf("Authorized on account %s", gBot.Self.UserName)

	//VARS
	db = storage.NewPGStorage(context.Background(), connStr)
	cloud = disk.New()

	acnt = accountant.NewAccountant(db)
	sync = synchronizer.New(cloud, db)

	//bot user
	os.Setenv("BOT_ADMIN", BOT_ADMIN)
	NewBotUser(BOT_ADMIN)
	//init done
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = CONFIG_TIMEOUT
	updates := gBot.GetUpdatesChan(updateConfig)

	gBot.Send(initCommands())

	for update := range updates {
		if !checkUser(update.SentFrom().UserName) {
			continue
		}

		if update.CallbackQuery != nil {
			callbackQueryHandler(update.CallbackQuery)
			continue
		}

		if update.Message != nil {
			if ResponseIsAwaited(update.SentFrom().UserName) {
				processResponse(&update)
			}

			// if update.Message.ReplyToMessage != nil {
			// 	processReply(&update)
			// }

			if update.Message.IsCommand() {
				processCommand(&update)
			}

			if len(update.Message.Text) > 0 {
				if _, err := strconv.Atoi(update.Message.Text); err == nil {
					processNumber(&update)
				}
			}
		}

	}

	return
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
