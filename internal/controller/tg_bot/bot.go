package tg_bot

import (
	"context"
	"finance-tg-bot/internal/entity"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/labstack/echo/v4"
)

type accountant interface {
	GetCatsLimit(ctx context.Context, user_id int) (cats []entity.TransCatLimit, err error)
	GetSubCats(ctx context.Context, user_id int, trans_cat string) (cats []string, err error)
	GetUserStatus(ctx context.Context, username string) (id int, status bool, err error)
	PostDoc(ctx context.Context, doc *entity.Document) (err error)
	DeleteDoc(chat_id, msg_id string, user_id int) (err error)
	MigrateFromCloud(ctx context.Context, username string) (err error)
	PushToCloud(ctx context.Context, username string) (err error)
	GetStatement(p map[string]string) (res string, err error)
	EditCats(ctx context.Context, tc *entity.TransCatLimit) (err error)
	Money2Time(transAmount int, user_id int) (res string, err error)
}

type Bot struct {
	api *tgbotapi.BotAPI
	accountant
	log *slog.Logger
}

type userChat struct {
	chatID    int64
	messageID int
	userName  string
	msgText   string
}

func New(api *tgbotapi.BotAPI, acc accountant, l *slog.Logger) *Bot {
	return &Bot{
		api:        api,
		accountant: acc,
		log:        l,
	}
}

func (b *Bot) deleteMsg(chatID int64, messageID int) {
	b.api.Send(tgbotapi.NewDeleteMessage(chatID, messageID))
}

func (b *Bot) clearMsgReplyMarkup(chatID int64, messageID int) {
	mrkp := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0),
	}

	b.api.Send(tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, mrkp))
}

func (b *Bot) updateMsgText(chatID int64, messageID int, text string) {
	b.api.Send(tgbotapi.NewEditMessageText(chatID, messageID, text))
}

func (b *Bot) requestCats(ctx context.Context, page int, uc *userChat) {
	cats, err := b.accountant.GetCatsLimit(ctx, BotUsers[uc.userName].UserId)
	if err != nil {
		b.log.Error("requestCats GetCatsLimit", "err", err)
		return
	}
	options := make([][]string, len(cats))
	for i, v := range cats {
		options[i] = []string{
			v.BalanceText,
			fmt.Sprintf("%s:%s", PREFIX_CATEGORY, v.Category),
		}
	}

	resMrkp, err := newKeyboardForm().setOptions(options).
		addNavigationControl(page, []string{EMOJI_CROSS, fmt.Sprintf("%s:%s", PREFIX_CATEGORY, "cancel")}, nil).
		getMarkup()
	if err != nil {
		b.log.Error("requestCats getMarkup", "err", err)
		return
	}

	if uc.msgText == "" {
		b.api.Send(tgbotapi.NewEditMessageReplyMarkup(uc.chatID, uc.messageID, resMrkp))
	} else {
		msg := tgbotapi.NewMessage(uc.chatID, uc.msgText)
		msg.ReplyMarkup = resMrkp
		b.api.Send(msg)
	}
}

func (b *Bot) processNumber(ctx context.Context, u *tgbotapi.Update) (err error) {
	b.requestCats(ctx, 0,
		&userChat{u.Message.Chat.ID, u.Message.MessageID, u.SentFrom().UserName, u.Message.Text + "₽"})
	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID)

	return err
}

func (b *Bot) confirmRecord(q *tgbotapi.CallbackQuery) {
	b.clearMsgReplyMarkup(q.Message.Chat.ID, q.Message.MessageID)

	finMsg, err := NewFinMsg().parseFinMsg(q.Message.Text)
	if err != nil {
		b.log.Error("confirmRecord parseFinMsg", "err", err)
		b.api.Send(tgbotapi.NewMessage(q.Message.Chat.ID, "error: something went wrong with confirmRecord:strconv.Atoi "+err.Error()))
		return
	}

	doc := &entity.Document{
		RecTime:     time.Unix(int64(q.Message.Date), 0),
		Category:    finMsg.category,
		Amount:      int64(finMsg.amount),
		Description: finMsg.description,
		MsgID:       fmt.Sprint(q.Message.MessageID),
		ChatID:      fmt.Sprint(q.Message.Chat.ID),
		UserId:      BotUsers[q.From.UserName].UserId,
	}

	err = b.accountant.PostDoc(context.Background(), doc)
	if err != nil {
		b.api.Send(tgbotapi.NewMessage(q.Message.Chat.ID, "error: something went wrong with confirmRecord:b.accountant.PostDoc "+err.Error()))
	}
}

func (b *Bot) deleteRecord(q *tgbotapi.CallbackQuery) {
	err := b.accountant.DeleteDoc(fmt.Sprint(q.Message.Chat.ID), fmt.Sprint(q.Message.MessageID), BotUsers[q.From.UserName].UserId)
	if err != nil {
		b.log.Error("deleteRecord DeleteDoc", "err", err)
		b.api.Send(tgbotapi.NewMessage(q.Message.Chat.ID, "error: something went wrong with deleteRecord:b.accountant.DeleteDoc "+err.Error()))
		return
	}
	b.deleteMsg(q.Message.Chat.ID, q.Message.MessageID)
}

func (b *Bot) requestReply(q *tgbotapi.CallbackQuery, respCode string) {
	b.clearMsgReplyMarkup(q.Message.Chat.ID, q.Message.MessageID)
	msg := tgbotapi.NewMessage(q.Message.Chat.ID, EMOJI_COMMENT+"...")
	msg.ReplyMarkup = getReply()
	b.api.Send(msg)
	waitUserResponeStart(b.log, q.From.UserName, respCode, *q.Message)
}

func (b *Bot) requestSubCats(ctx context.Context, page int, q *tgbotapi.CallbackQuery) {
	var cat string

	if len(strings.Split(q.Message.Text, " ")) < 3 {
		cat, _ = strings.CutPrefix(q.Data, PREFIX_CATEGORY+":")
	} else {
		cat = strings.Join(strings.Split(q.Message.Text, " ")[2:], " ")
	}

	subCats, err := b.accountant.GetSubCats(ctx, BotUsers[q.From.UserName].UserId, cat)
	if err != nil {
		b.log.Error("requestSubCats GetSubCats", "err", err)
		return
	}

	options := make([][]string, len(subCats))
	for i, v := range subCats {
		options[i] = []string{v, fmt.Sprintf("%s:%s:%d", PREFIX_SUBCATEGORY, cat, i)}
	}

	resMrkp, err := newKeyboardForm().setOptions(options).
		addNavigationControl(page,
			[]string{EMOJI_HOOK_BACK, fmt.Sprintf("%s:%s", PREFIX_SUBCATEGORY, "backToCategories")},
			[]string{EMOJI_KEYBOARD, fmt.Sprintf("%s:%s", PREFIX_SUBCATEGORY, "writeCustom")}).
		getMarkup()
	if err != nil {
		b.log.Error("requestSubCats mrkp.getMarkup", "err", err)
		return
	}

	b.api.Send(tgbotapi.NewEditMessageReplyMarkup(q.Message.Chat.ID, q.Message.MessageID, resMrkp))
}

func (b *Bot) handleUpdate(ctx context.Context, u *tgbotapi.Update) {
	if !b.checkUser(b.log, ctx, u.SentFrom().UserName) {
		b.log.Info("handleUpdate:checkUser is false - skip update")
		return
	}

	if u.CallbackQuery != nil {
		b.callbackQueryHandler(ctx, u.CallbackQuery)
		return
	}

	if u.Message != nil {
		if responseIsAwaited(u.SentFrom().UserName) {
			b.responseHandler(ctx, u)
			return
		} else if u.Message.ReplyToMessage != nil {
			b.replyHandler(u)
			return
		}

		if u.Message.IsCommand() {
			b.processCommand(ctx, u)
			return
		}

		if len(u.Message.Text) > 0 {
			if _, err := strconv.Atoi(u.Message.Text); err == nil {
				b.processNumber(ctx, u)
				return
			}
		}
	}
}

func (b *Bot) Run(ctx context.Context, port string) (err error) {
	wh, err := b.api.GetWebhookInfo()
	if err != nil {
		b.log.Error("bot.Run GetWebhookInfo", "err", err)
		return err
	}
	b.log.Info("webhook info", "whInfo", fmt.Sprintf("%#v", wh))

	if wh.URL == "" {
		updateConfig := tgbotapi.NewUpdate(0)
		updateConfig.Timeout = 60

		updates := b.api.GetUpdatesChan(updateConfig)

		b.api.Send(initCommands())

		for {
			select {
			case update := <-updates:
				ctxU, cancelU := context.WithTimeout(ctx, 7*time.Second)
				b.handleUpdate(ctxU, &update)
				cancelU()
			case <-ctx.Done():
				b.api.StopReceivingUpdates()
				return ctx.Err()
			}
		}
	} else {
		if wh.LastErrorDate != 0 {
			b.log.Info("telegram last error message", "date", time.Unix(int64(wh.LastErrorDate), 0), "msg", wh.LastErrorMessage)
		}

		b.api.Send(initCommands())

		e := echo.New()
		e.POST("/", b.requestHandler)
		e.Start(":" + port)
	}

	return
}

func (b *Bot) requestHandler(c echo.Context) error {
	defer func() {
		if r := recover(); r != nil {
			b.log.Error("requestHandler recovered panic", "r", r)
		}
	}()

	var update tgbotapi.Update
	if err := c.Bind(&update); err != nil {
		b.log.Warn("cannot bind update", "err", err)
		return c.JSON(204, nil)
	}

	b.handleUpdate(context.Background(), &update)

	return c.JSON(200, nil)
}
