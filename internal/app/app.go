package app

import (
	"context"
	"log/slog"

	"os"
	"os/signal"
	"syscall"

	"finance-tg-bot/config"
	tg_bot "finance-tg-bot/internal/controller/tg_bot"
	accountant "finance-tg-bot/internal/usecase"
	cloud "finance-tg-bot/internal/usecase/cloud"
	repo "finance-tg-bot/internal/usecase/repo"
	reports "finance-tg-bot/internal/usecase/repo/reports"
	users "finance-tg-bot/internal/usecase/repo/users"
	"finance-tg-bot/pkg/postgres"
	"finance-tg-bot/pkg/repository"
	"finance-tg-bot/pkg/ydb"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Run(config config.Config) (err error) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	gBot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		log.Error("app - Run - tgbotapi.NewBotAPI", "err", err)
		return
	}
	gBot.Debug = true
	log.Info("authorized on account", "botName", gBot.Self.UserName)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	postgres, err := postgres.New(ctx, config.DatabaseDSN)
	if err != nil {
		log.Error("app - Run - postgres.New", "err", err)
		return
	}
	defer postgres.Close()

	ydb, err := ydb.NewNative(ctx,
		config.YdbDSN,
		"authorized_key.json")
	if err != nil {
		log.Error("app - Run - ydb.New", "err", err)
		return
	}
	defer ydb.Close(ctx)

	acnt := accountant.New(
		repo.New(postgres, ydb),
		users.New(*ydb),
		reports.New(&repository.Repository{Postgres: postgres, Ydb: ydb}),
		cloud.New(),
		log,
	)
	bot := tg_bot.New(gBot, acnt)

	return bot.Run(ctx)
}
