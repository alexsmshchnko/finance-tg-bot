package main

import (
	"finance-tg-bot/internal"
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	gToken  string
	gChatID int64
	gBot    *tgbotapi.BotAPI
)

func init() {
	if gToken = os.Getenv(BOT_TOKEN_NAME); gToken == "" {
		panic(fmt.Errorf("failed to load env variable %s", BOT_TOKEN_NAME))
	}

	var err error
	if gBot, err = tgbotapi.NewBotAPI(gToken); err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", gBot.Self.UserName)
}

func sendMsg(msg string) error {
	_, err := gBot.Send(tgbotapi.NewMessage(gChatID, msg))
	return err
}

func run() error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = CONFIG_TIMEOUT

	//str := fmt.Sprint(internal.GetExpenseCategories())

	for update := range gBot.GetUpdatesChan(updateConfig) {
		if update.Message != nil {
			gChatID = update.Message.Chat.ID

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyMarkup = getCategoryKeyboard(internal.GetExpenseCategories())
			gBot.Send(msg)

			//sendMsg(str)
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal()
	}

}
