package main

import (
	"context"
	"log"

	"os"
	"os/signal"
	"syscall"

	"finance-tg-bot/internal/accountant"
	"finance-tg-bot/internal/config"
	"finance-tg-bot/internal/disk"
	"finance-tg-bot/internal/storage"
	"finance-tg-bot/internal/synchronizer"
	"finance-tg-bot/internal/tg_bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func run() (err error) {
	gBot, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Println("[ERROR] failed to create botAPI")
		return
	}
	log.Printf("Authorized on account %s", gBot.Self.UserName)
	gBot.Debug = true

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db := storage.NewPGStorage(ctx, config.Get().DatabaseDSN)
	cloud := disk.New()
	acnt := accountant.NewAccountant(db)
	sync := synchronizer.New(cloud, db)
	bot := tg_bot.New(gBot, acnt, sync)

	return bot.Run(ctx)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
