package app

import (
	"context"
	"fmt"
	"log"

	"os"
	"os/signal"
	"syscall"

	"finance-tg-bot/config"
	tg_bot "finance-tg-bot/internal/controller/tg_bot"
	accountant "finance-tg-bot/internal/usecase"
	cloud "finance-tg-bot/internal/usecase/cloud"
	repo "finance-tg-bot/internal/usecase/repo"
	"finance-tg-bot/pkg/postgres"

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

	postgres, err := postgres.New(ctx, config.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("app - Run - postgres.New: %w", err)
	}
	defer postgres.Close()

	acnt := accountant.New(repo.New(postgres), cloud.New())
	bot := tg_bot.New(gBot, acnt)

	return bot.Run(ctx)
}
