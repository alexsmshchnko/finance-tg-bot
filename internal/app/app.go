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
	doc "finance-tg-bot/internal/usecase/repo"
	report "finance-tg-bot/internal/usecase/repo/reports"
	user "finance-tg-bot/internal/usecase/repo/users"
	"finance-tg-bot/pkg/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Run(config config.Config) (err error) {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	log.Info("configuration", "config.ServerPort", config.ServerPort)

	gBot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		log.Error("app - Run - tgbotapi.NewBotAPI", "err", err)
		return
	}
	gBot.Debug = true
	log.Info("authorized on account", "botName", gBot.Self.UserName)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	r := repository.NewRepository(
		config.RepoServiceDomain,
		config.RepoAuthToken,
		log,
	)

	acnt := accountant.New(
		doc.New(r), user.New(r), report.New(r),
		cloud.New(), log)
	bot := tg_bot.New(gBot, acnt)

	return bot.Run(ctx, config.ServerPort)
}
