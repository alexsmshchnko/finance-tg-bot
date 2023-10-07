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
	if gToken = os.Getenv("BOT_TOKEN"); gToken == "" {
		panic(fmt.Errorf("failed to load env variable %s", "BOT_TOKEN"))
	}

	var err error
	if gBot, err = tgbotapi.NewBotAPI(gToken); err != nil {
		log.Panic(err)
	}
}

func run() error {
	//fmt.Println(gBot)

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal()
	}

}
