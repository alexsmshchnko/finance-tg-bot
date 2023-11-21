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

func processNumber(u *tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, u.Message.Text+"₽")
	cats, err := acnt.GetCats(context.Background(), u.Message.Chat.UserName)
	if err != nil {
		return err
	}
	msg.ReplyMarkup = getCatInlineKeyboard(cats)
	_, err = gBot.Send(msg)
	deleteMsg(u.Message.Chat.ID, u.Message.MessageID)

	return err
}

func processCallbackCategory(u *tgbotapi.Update) (err error) {
	cat, _ := strings.CutPrefix(u.CallbackQuery.Data, "CAT:")
	u.CallbackQuery.Message.Text, _ = strings.CutSuffix(u.CallbackQuery.Message.Text, "₽")
	//amnt, _ := strconv.Atoi(u.CallbackQuery.Message.Text)

	resp := u.CallbackQuery.Message.Text + "₽ на " + cat
	msg := tgbotapi.NewMessage(u.CallbackQuery.Message.Chat.ID, resp)
	msg.ReplyMarkup = getMsgOptionsKeyboard()
	_, err = gBot.Send(msg)
	deleteMsg(u.CallbackQuery.Message.Chat.ID, u.CallbackQuery.Message.MessageID)

	// rec := internal.NewReceiptRec(time.Now(), cat, amnt, "")
	// internal.AddNewExpense(rec)

	return
}

func confirmRecord(u *tgbotapi.Update) {
	clearMsgReplyMarkup(u.CallbackQuery.Message.Chat.ID, u.CallbackQuery.Message.MessageID)
}

func deleteRecord(u *tgbotapi.Update) {
	acnt.DeleteDoc(fmt.Sprint(u.CallbackQuery.Message.MessageID), u.CallbackQuery.From.UserName)
	deleteMsg(u.CallbackQuery.Message.Chat.ID, u.CallbackQuery.Message.MessageID)
}

// func addDescription(u *tgbotapi.Update) (err error) {
// 	rec := internal.ReceiptRec{Description: u.Message.Text}
// 	err = internal.AddLastExpenseDescription(&rec)
// 	if err != nil {
// 		return
// 	}

// 	updateMsgText(u.Message.Chat.ID, u.Message.ReplyToMessage.MessageID, u.Message.ReplyToMessage.Text+"\n"+EMOJI_COMMENT+u.Message.Text)

// 	deleteMsg(u.Message.Chat.ID, u.Message.MessageID)

// 	return
// }

func requestDescription(u *tgbotapi.Update) {
	// msg := tgbotapi.NewMessage(u.CallbackQuery.Message.Chat.ID, "описание")
	// msg.ReplyMarkup = getReply()

	msg := tgbotapi.NewMessage(u.CallbackQuery.Message.Chat.ID, EMOJI_COMMENT+"...")

	msg.ReplyMarkup = getReply()
	gBot.Send(msg)
	WaitUserResponeStart(u.SentFrom().UserName, "REC_DESC", *u.CallbackQuery.Message)
}

func processCallbackOption(u *tgbotapi.Update) (err error) {
	var r string

	switch u.CallbackQuery.Data {
	case "OPT:saveRecord":
		confirmRecord(u)
	case "OPT:addDescription":
		requestDescription(u)
	case "OPT:deleteRecord":
		deleteRecord(u)
	}

	msg := tgbotapi.NewMessage(u.CallbackQuery.Message.Chat.ID, r)
	_, err = gBot.Send(msg)

	return
}

func processCallback(u *tgbotapi.Update) (err error) {
	if strings.Contains(u.CallbackQuery.Data, "CAT:") {
		err = processCallbackCategory(u)
	} else if strings.Contains(u.CallbackQuery.Data, "OPT:") {
		err = processCallbackOption(u)
	}
	return
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
	// 	return
	// }

	respMsg := BotUsers[u.SentFrom().UserName].ResponseMsg
	updateMsgText(u.Message.Chat.ID, respMsg.MessageID, respMsg.Text+"\n"+EMOJI_COMMENT+u.Message.Text)

	amnt, _ := strconv.Atoi(strings.Split(respMsg.Text, "₽")[0])
	cat := strings.Split(respMsg.Text, " на ")[1]

	//	acnt.
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
			if update.CallbackQuery.Data != "" {
				processCallback(&update)
			}
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
