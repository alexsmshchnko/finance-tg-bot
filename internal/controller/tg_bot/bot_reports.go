package tg_bot

import (
	"context"
	"finance-tg-bot/internal/entity"
	"fmt"
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
	case "deleteReport":
		b.deleteMsg(query.Message.Chat.ID, query.Message.MessageID)
	case "saveReport":
		b.clearMsgReplyMarkup(query.Message.Chat.ID, query.Message.MessageID)
	}

}

func getReportKeyboard() *tgbotapi.InlineKeyboardMarkup {
	mrkp := newKeyboardForm()
	mrkp.setOptions([][]string{
		{"День", PREFIX_REPORT + ":monthReport:day"},
		// {"Неделя", PREFIX_REPORT + ":monthReport:week"},
		{"Текущий месяц", PREFIX_REPORT + ":monthReport:current"},
		{"Предыдущий месяц", PREFIX_REPORT + ":monthReport:previous"},
	})
	mrkp.setControl([][][]string{
		{{EMOJI_CROSS, PREFIX_REPORT + ":cancelReport"}},
	})
	res, err := mrkp.getMarkup()
	if err != nil {
		return nil
	}
	return res
}

func statementReport(b *Bot, q *tgbotapi.CallbackQuery) {
	var (
		t, t2 time.Time
		text  string
		err   error
	)
	t = time.Now()

	switch strings.Split(q.Data, ":")[2] {
	case "day":
		t2 = t.AddDate(0, 0, 1).Add(-time.Second)
	case "current":
		t2 = t.AddDate(0, 0, 1).Add(-time.Second)
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	case "previous":
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).AddDate(0, -1, 0)
		t2 = t.AddDate(0, 1, 0).Add(-time.Second)
	}

	rep := &entity.Report{
		RepName: "TotalsForThePeriod",
		RepParms: map[string]string{
			"username": q.From.UserName,
			"datefrom": t.Format("02.01.2006"),
			"dateto":   t2.Format("02.01.2006"),
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

	mrkp := newKeyboardForm()
	mrkp.setControl([][][]string{
		{
			{EMOJI_CROSS, PREFIX_REPORT + ":deleteReport"},
			{EMOJI_SAVE, PREFIX_REPORT + ":saveReport"},
		},
	})
	resMrkp, err := mrkp.getMarkup()
	if err != nil {
		fmt.Println(err)
		return
	}
	ms := tgbotapi.NewEditMessageReplyMarkup(q.Message.Chat.ID, q.Message.MessageID, *resMrkp)
	b.api.Send(ms)
}
