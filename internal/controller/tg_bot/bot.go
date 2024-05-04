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

	msg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, mrkp)
	b.api.Send(msg)
}

func (b *Bot) updateMsgText(chatID int64, messageID int, text string) {
	b.api.Send(tgbotapi.NewEditMessageText(chatID, messageID, text))
}

func (b *Bot) requestCats(ctx context.Context, page int, q *tgbotapi.CallbackQuery, u *tgbotapi.Update) {
	var (
		userName  string
		chatID    int64
		messageID int
		limit     string
		direction string
	)

	if u == nil {
		userName = q.From.UserName
		chatID = q.Message.Chat.ID
		messageID = q.Message.MessageID
	} else {
		userName = u.SentFrom().UserName
		chatID = u.Message.Chat.ID
		messageID = u.Message.MessageID
	}

	// TODO move formatting to usecase (no only here)
	cats, err := b.accountant.GetCatsLimit(ctx, BotUsers[userName].UserId)
	if err != nil {
		return
	}
	options := make([][]string, len(cats))
	for i, v := range cats {
		switch v.Direction {
		case -1:
			direction = EMOJI_DEBIT
		case 0:
			direction = EMOJI_DEPOSIT
		case 1:
			direction = EMOJI_CREDIT
		}
		if v.Limit > 0 {
			limit = fmt.Sprintf(" (%d)", v.Balance)
		} else {
			limit = ""
		}
		options[i] = []string{
			v.Category + " " + direction + limit,
			PREFIX_CATEGORY + ":" + v.Category,
		}
	}

	mrkp := newKeyboardForm()
	mrkp.setOptions(options)
	mrkp.addNavigationControl(page, nil, nil)
	resMrkp, err := mrkp.getMarkup()
	if err != nil {
		fmt.Println(err)
		return
	}

	if u == nil {
		msg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, *resMrkp)
		b.api.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, u.Message.Text+"₽")
		msg.ReplyMarkup = resMrkp
		b.api.Send(msg)
	}
}

func (b *Bot) processNumber(ctx context.Context, u *tgbotapi.Update) (err error) {
	b.requestCats(ctx, 0, nil, u)
	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID)

	return err
}

func (b *Bot) confirmRecord(q *tgbotapi.CallbackQuery) {
	b.clearMsgReplyMarkup(q.Message.Chat.ID, q.Message.MessageID)
	var (
		err        error
		amnt       int
		cat, descr string
	)
	amnt, err = strconv.Atoi(strings.Split(q.Message.Text, "₽")[0])
	if err != nil {
		b.api.Send(tgbotapi.NewMessage(q.Message.Chat.ID, "error: something went wrong with confirmRecord:strconv.Atoi "+err.Error()))
		return
	}
	scntSplit := strings.Split(q.Message.Text, "\n")
	cat = strings.Split(scntSplit[0], " на ")[1]
	if len(scntSplit) > 1 {
		descr, _ = strings.CutPrefix(scntSplit[1], EMOJI_COMMENT)
	}

	doc := &entity.Document{
		RecTime:     time.Unix(int64(q.Message.Date), 0),
		Category:    cat,
		Amount:      int64(amnt),
		Description: descr,
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
	b.accountant.DeleteDoc(fmt.Sprint(q.Message.Chat.ID), fmt.Sprint(q.Message.MessageID), BotUsers[q.From.UserName].UserId)
	b.deleteMsg(q.Message.Chat.ID, q.Message.MessageID)
}

// func addDescription(u *tgbotapi.Update) (err error) {
// 	rec := internal.ReceiptRec{Description: u.Message.Text}
// 	err = internal.AddLastExpenseDescription(&rec)
// 	if err != nil {
//

// 	updateMsgText(u.Message.Chat.ID, u.Message.ReplyToMessage.MessageID, u.Message.ReplyToMessage.Text+"\n"+EMOJI_COMMENT+u.Message.Text)

// 	deleteMsg(u.Message.Chat.ID, u.Message.MessageID)

//

// func (b *Bot) requestDescription(query *tgbotapi.CallbackQuery) {
// 	subCat := strings.Join(strings.Split(query.Message.Text, " ")[2:], " ")
// 	cats, _ := b.accountant.GetSubCats(context.Background(), query.From.UserName, subCat)
// 	mrkp := getPagedListInlineKeyboard(cats, 0, PREFIX_SUBCATEGORY, PREFIX_SUBCATEGORY+":"+EMOJI_KEYBOARD)
// 	msg := tgbotapi.NewEditMessageReplyMarkup(query.Message.Chat.ID, query.Message.MessageID, *mrkp)
// 	b.api.Send(msg)
// }

func (b *Bot) requestReply(q *tgbotapi.CallbackQuery, respCode string) {
	b.clearMsgReplyMarkup(q.Message.Chat.ID, q.Message.MessageID)
	msg := tgbotapi.NewMessage(q.Message.Chat.ID, EMOJI_COMMENT+"...")
	msg.ReplyMarkup = getReply()
	b.api.Send(msg)
	waitUserResponeStart(q.From.UserName, respCode, *q.Message)
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
	if len(subCats) == 0 {
		subCats = []string{" "}
	}
	options := make([][]string, len(subCats))
	for i, v := range subCats {
		options[i] = []string{v, fmt.Sprintf("%s:%s:%d", PREFIX_SUBCATEGORY, cat, i)}
	}

	mrkp := newKeyboardForm()
	mrkp.setOptions(options)
	mrkp.addNavigationControl(page, nil, []string{EMOJI_KEYBOARD, fmt.Sprintf("%s:%s", PREFIX_SUBCATEGORY, "writeCustom")})
	resMrkp, err := mrkp.getMarkup()
	if err != nil {
		b.log.Error("requestSubCats mrkp.getMarkup", "err", err)
		return
	}

	msg := tgbotapi.NewEditMessageReplyMarkup(q.Message.Chat.ID, q.Message.MessageID, *resMrkp)
	b.api.Send(msg)
}

// func processReply(u *tgbotapi.Update) (err error) {
// 	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "")
// 	msg.ReplyMarkup = getReply()

// 	gBot.Send(msg)

// 	return
// }

func (b *Bot) handleUpdate(ctx context.Context, u *tgbotapi.Update) {
	if !b.checkUser(ctx, u.SentFrom().UserName) {
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
		}

		// if update.Message.ReplyToMessage != nil {
		// 	processReply(&update)
		// }

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
				ctxU, cancelU := context.WithTimeout(ctx, 3*time.Second)
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
			fmt.Println("recovered panic", r)
		}
	}()

	var update tgbotapi.Update
	if err := c.Bind(&update); err != nil {
		fmt.Println("Cannot bind update", err)
		return c.JSON(204, nil)
	}

	b.handleUpdate(context.Background(), &update)

	return c.JSON(200, nil)
}
