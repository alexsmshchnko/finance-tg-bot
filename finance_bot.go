package main

import (
	"finance-tg-bot/internal"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	gToken string
	gBot   *tgbotapi.BotAPI
)

func init() {
	if gToken = os.Getenv(BOT_TOKEN_NAME); gToken == "" {
		panic(fmt.Errorf("failed to load env variable %s", BOT_TOKEN_NAME))
	}

	var err error
	if gBot, err = tgbotapi.NewBotAPI(gToken); err != nil {
		log.Panic(err)
	}
	gBot.Debug = true

	log.Printf("Authorized on account %s", gBot.Self.UserName)
}

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
	msg.ReplyMarkup = getCatInlineKeyboard(internal.GetExpenseCategories())
	_, err := gBot.Send(msg)
	deleteMsg(u.Message.Chat.ID, u.Message.MessageID)

	return err
}

func processCallbackCategory(u *tgbotapi.Update) (err error) {
	cat, _ := strings.CutPrefix(u.CallbackQuery.Data, "CAT:")
	u.CallbackQuery.Message.Text, _ = strings.CutSuffix(u.CallbackQuery.Message.Text, "₽")
	amnt, _ := strconv.Atoi(u.CallbackQuery.Message.Text)

	resp := u.CallbackQuery.Message.Text + "₽ на " + cat
	msg := tgbotapi.NewMessage(u.CallbackQuery.Message.Chat.ID, resp)
	msg.ReplyMarkup = getMsgOptionsKeyboard()
	_, err = gBot.Send(msg)
	deleteMsg(u.CallbackQuery.Message.Chat.ID, u.CallbackQuery.Message.MessageID)

	rec := internal.NewReceiptRec(time.Now(), cat, amnt, "")
	internal.AddNewExpense(rec)

	return
}

func confirmRecord(u *tgbotapi.Update) {
	clearMsgReplyMarkup(u.CallbackQuery.Message.Chat.ID, u.CallbackQuery.Message.MessageID)
}

func deleteRecord(u *tgbotapi.Update) {
	internal.DeleteLastExpense()
	deleteMsg(u.CallbackQuery.Message.Chat.ID, u.CallbackQuery.Message.MessageID)
}

func addDescription(u *tgbotapi.Update) (err error) {
	rec := internal.ReceiptRec{Description: u.Message.Text}
	err = internal.AddLastExpenseDescription(&rec)
	if err != nil {
		return
	}

	updateMsgText(u.Message.Chat.ID, u.Message.ReplyToMessage.MessageID, u.Message.ReplyToMessage.Text+"\n"+EMOJI_COMMENT+u.Message.Text)

	deleteMsg(u.Message.Chat.ID, u.Message.MessageID)

	return
}

func requestReply(u *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(u.CallbackQuery.Message.Chat.ID, "описание")
	msg.ReplyMarkup = getReply()

	gBot.Send(msg)
}

func processCallbackOption(u *tgbotapi.Update) (err error) {
	var r string

	switch u.CallbackQuery.Data {
	case "OPT:saveRecord":
		confirmRecord(u)
	case "OPT:addDescription":
		requestReply(u)
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

func processReply(u *tgbotapi.Update) (err error) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "")
	msg.ReplyMarkup = getReply()

	gBot.Send(msg)

	return
}

func run() error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = CONFIG_TIMEOUT
	updates := gBot.GetUpdatesChan(updateConfig)

	gBot.Send(initCommands())

	for update := range updates {
		if update.CallbackQuery != nil {
			if update.CallbackQuery.Data != "" {
				processCallback(&update)
			}
		}

		if update.Message != nil {
			if update.Message.ReplyToMessage != nil {
				processReply(&update)
			}

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

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal()
	}

}
