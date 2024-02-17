package tg_bot

import (
	"context"
	"finance-tg-bot/internal/entity"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleReportCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) {
	split := strings.Split(query.Data, ":")
	switch split[1] {
	case "monthReport":
		statementReport(b, query)
	case "cancelReport":
		b.deleteMsg(query.Message.Chat.ID, query.Message.MessageID)
	}
}

func statementReport(b *Bot, q *tgbotapi.CallbackQuery) {
	var (
		t    time.Time
		text string
		err  error
	)
	t = time.Now()

	switch strings.Split(q.Data, ":")[2] {
	case "current":
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	case "previous":
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).AddDate(0, -1, 0)
	}

	rep := &entity.Report{
		RepName: "TotalsForThePeriod",
		RepParms: map[string]string{
			"username": q.From.UserName,
			"datefrom": t.Format("02.01.2006"),
			"dateto":   t.AddDate(0, 1, 0).Add(-time.Second).Format("02.01.2006"),
		},
	}

	text, err = b.accountant.GetStatement(rep)
	if err != nil {
		return
	}
	text = "*" + t.Format("January 2006") + "*\n```\n" + text + "\n" + "```"

	msg := tgbotapi.NewEditMessageText(q.Message.Chat.ID, q.Message.MessageID, text)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}
