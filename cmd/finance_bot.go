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
	var (
		gBot  *tgbotapi.BotAPI
		db    *storage.PGStorage
		cloud *disk.Disk
		acnt  *accountant.Accountant
		sync  *synchronizer.Synchronizer
		bot   *tg_bot.Bot

		// log    *slog.Logger
		ctx    context.Context
		cancel context.CancelFunc
	)

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

	bot = tg_bot.New(gBot, acnt, sync)

	err = bot.Run(ctx)

	return
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
