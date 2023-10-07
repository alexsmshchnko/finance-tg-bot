package main

import (
	"fmt"
	"log"
	"os"

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
	log.Printf("Authorized on account %s", gBot.Self.UserName)
}

func run() error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = CONFIG_TIMEOUT

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal()
	}

}
