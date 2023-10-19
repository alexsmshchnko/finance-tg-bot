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

func processCommand(u *tgbotapi.Update) (err error) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "")

	// Extract the command from the Message.
	switch u.Message.Command() {
	case "help":
		msg.Text = "I understand /sayhi and /status."
	case "start":
		msg.Text = "Hi :)"
	case "sync":
		err = internal.SyncDiskFile(u.Message.Chat.UserName)
		msg.Text = "In progress"
	default:
		msg.Text = "I don't know that command"
	}
	if err != nil {
		log.Println(err)
		return
	}

	_, err = gBot.Send(msg)

	return
}

func processNumber(u *tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, u.Message.Text+"₽")
	msg.ReplyMarkup = getCategoryKeyboard(internal.GetExpenseCategories())
	_, err := gBot.Send(msg)

	return err
}

func processCallbackCategory(u *tgbotapi.Update) (err error) {
	cat, _ := strings.CutPrefix(u.CallbackQuery.Data, "CAT:")
	u.CallbackQuery.Message.Text, _ = strings.CutSuffix(u.CallbackQuery.Message.Text, "₽")
	amnt, _ := strconv.Atoi(u.CallbackQuery.Message.Text)

	resp := u.CallbackQuery.Message.Text + "₽ на категорию " + cat
	msg := tgbotapi.NewMessage(u.CallbackQuery.Message.Chat.ID, resp)
	_, err = gBot.Send(msg)
	deleteMsg(u.CallbackQuery.Message.Chat.ID, u.CallbackQuery.Message.MessageID)

	rec := internal.NewReceiptRec(time.Now(), cat, amnt, "")
	internal.AddNewExpense(rec)

	return
}

func processCallback(u *tgbotapi.Update) (err error) {
	if strings.Contains(u.CallbackQuery.Data, "CAT:") {
		err = processCallbackCategory(u)
	}
	return
}

func run() error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = CONFIG_TIMEOUT

	updates := gBot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.CallbackQuery != nil {
			if update.CallbackQuery.Data != "" {
				processCallback(&update)
			}
		}

		if update.Message != nil {
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
