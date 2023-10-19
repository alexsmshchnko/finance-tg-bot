package main

import (
	"finance-tg-bot/internal"
	"fmt"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	gToken string
	//gChatID int64
	gBot *tgbotapi.BotAPI
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

func processCommand(u *tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "")

	// Extract the command from the Message.
	switch u.Message.Command() {
	case "help":
		msg.Text = "I understand /sayhi and /status."
	case "start":
		msg.Text = "Hi :)"
	case "status":
		msg.Text = "I'm ok."
	default:
		msg.Text = "I don't know that command"
	}

	_, err := gBot.Send(msg)

	return err
}

func processNumber(u *tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, u.Message.Text)
	msg.ReplyMarkup = getCategoryKeyboard(internal.GetExpenseCategories())
	_, err := gBot.Send(msg)

	return err
}

func run() error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = CONFIG_TIMEOUT

	updates := gBot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if update.Message.IsCommand() {
			processCommand(&update)
		}

		if update.Message != nil {
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
