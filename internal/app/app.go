package app

import (
	"context"
	"fmt"
	"log"

	"os"
	"os/signal"
	"syscall"

	"finance-tg-bot/config"
	"finance-tg-bot/internal/accountant"
	"finance-tg-bot/internal/disk"
	"finance-tg-bot/internal/storage"
	"finance-tg-bot/internal/synchronizer"
	"finance-tg-bot/internal/tg_bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Run(config config.Config) (err error) {
	gBot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		return fmt.Errorf("app - Run - tgbotapi.NewBotAPI: %w", err)
	}
	log.Printf("authorized on account %s", gBot.Self.UserName)
	gBot.Debug = true

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db, err := storage.New(ctx, config.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("app - Run - db.New: %w", err)
	}
	defer db.Close()

	cloud := disk.New()
	acnt := accountant.New(db)
	sync := synchronizer.New(cloud, db)
	bot := tg_bot.New(gBot, acnt, sync)

	return bot.Run(ctx)
}
