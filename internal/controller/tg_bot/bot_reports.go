package tg_bot

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleReportCallbackQuery(query *tgbotapi.CallbackQuery) {
	split := strings.Split(query.Data, ":")
	switch split[1] {
	case "monthReport":
		statementReport(b, query)
	case "hist":
		historyReport(b, query)
	case "cancelReport":
		b.deleteMsg(query.Message.Chat.ID, query.Message.MessageID)
	case "deleteReport":
		b.deleteMsg(query.Message.Chat.ID, query.Message.MessageID)
	case "saveReport":
		b.clearMsgReplyMarkup(query.Message.Chat.ID, query.Message.MessageID)
	}

}

func getReportKeyboard() (resMrkp tgbotapi.InlineKeyboardMarkup) {
	resMrkp, _ = newKeyboardForm().
		setOptions([][]string{
			{"День", PREFIX_REPORT + ":monthReport:day"},
			// {"Неделя", PREFIX_REPORT + ":monthReport:week"},
			{"Текущий месяц", PREFIX_REPORT + ":monthReport:current"},
			{"Текущий месяц с детализацией", PREFIX_REPORT + ":monthReport:curDet"},
			{"Предыдущий месяц", PREFIX_REPORT + ":monthReport:previous"},
		}).
		setControl([][][]string{
			{{EMOJI_CROSS, PREFIX_REPORT + ":cancelReport"}},
		}).
		getMarkup()

	return
}
func historyReport(b *Bot, q *tgbotapi.CallbackQuery) {
	fm, err := NewFinMsg().parseFinMsg(q.Message.Text)
	if err != nil {
		b.log.Error("historyReport failed on message parse", "err", err)
		return
	}

	t2 := time.Now()
	t := time.Date(t2.Year(), t2.Month(), 1, 0, 0, 0, 0, t2.Location()).AddDate(0, -6, 0)

	p := map[string]string{
		"User_id":     fmt.Sprint(BotUsers[q.From.UserName].UserId),
		"Report_type": strings.Split(q.Data, ":")[2],
		"Date_from":   t.Format("2006-01-02T15:04:05Z"),
		"Date_to":     t2.Format("2006-01-02T15:04:05Z"),
		"Add_attr1":   fm.category,
		"Add_attr2":   fm.description,
	}

	text, err := b.accountant.GetStatement(p)
	if err != nil {
		b.log.Error("historyReport GetStatement", "err", err)
		return
	}

	resMrkp, err := newKeyboardForm().
		setControl([][][]string{
			{
				{EMOJI_CROSS, PREFIX_REPORT + ":deleteReport"},
				{EMOJI_SAVE, PREFIX_REPORT + ":saveReport"},
			},
		}).
		getMarkup()
	if err != nil {
		b.log.Error("statementReport getMarkup", "err", err)
		return
	}
	msg := tgbotapi.NewMessage(q.Message.Chat.ID, "```\n"+text+"\n"+"```")
	msg.ReplyMarkup = &resMrkp
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

func statementReport(b *Bot, q *tgbotapi.CallbackQuery) {
	var (
		t, t2        time.Time
		i0           int
		header, text string
		err          error
	)
	t = time.Now().Truncate(time.Hour * 24)

	switch strings.Split(q.Data, ":")[2] {
	case "day":
		t2 = t.AddDate(0, 0, 1).Add(-time.Second)
		header = t.Format("02 January 2006")
	case "current":
		t2 = t.AddDate(0, 0, 1).Add(-time.Second)
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		header = t.Format("January 2006")
	case "curDet":
		t2 = t.AddDate(0, 0, 1).Add(-time.Second)
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		header = t.Format("January 2006")
	case "previous":
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).AddDate(0, -1, 0)
		t2 = t.AddDate(0, 1, 0).Add(-time.Second)
		header = t.Format("January 2006")
	}
	header = "*" + header + "*\n"

	i0 = BotUsers[q.From.UserName].UserId

	p := map[string]string{
		"User_id":   fmt.Sprint(i0),
		"Date_from": t.Format("2006-01-02T15:04:05Z"),
		"Date_to":   t2.Format("2006-01-02T15:04:05Z")}

	if strings.Split(q.Data, ":")[2] == "curDet" {
		p["Report_type"] = "subcat"
	}

	text, err = b.accountant.GetStatement(p)
	if err != nil {
		b.log.Error("statementReport GetStatement", "err", err)
		return
	}

	text = header + "```\n" + text + "\n" + "```"

	msg := tgbotapi.NewEditMessageText(q.Message.Chat.ID, q.Message.MessageID, text)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)

	resMrkp, err := newKeyboardForm().
		setControl([][][]string{
			{
				{EMOJI_CROSS, PREFIX_REPORT + ":deleteReport"},
				{EMOJI_SAVE, PREFIX_REPORT + ":saveReport"},
			},
		}).
		getMarkup()
	if err != nil {
		b.log.Error("statementReport getMarkup", "err", err)
		return
	}
	b.api.Send(tgbotapi.NewEditMessageReplyMarkup(q.Message.Chat.ID, q.Message.MessageID, resMrkp))
}
