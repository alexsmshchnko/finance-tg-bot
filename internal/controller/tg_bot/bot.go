package tg_bot

import (
	"context"
	"finance-tg-bot/internal/entity"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type accountant interface {
	GetCatsLimit(ctx context.Context, username, limit string) (cats []entity.TransCatLimit, err error)
	GetSubCats(ctx context.Context, username, trans_cat string) (cats []string, err error)
	GetUserStatus(ctx context.Context, username string) (status bool, err error)
	PostDoc(ctx context.Context, category string, amount int, description string, msg_id string, direction int, client string) (err error)
	DeleteDoc(msg_id string, client string) (err error)
	MigrateFromCloud(ctx context.Context, username string) (err error)
	PushToCloud(ctx context.Context, username string) (err error)
	GetStatement(p map[string]string) (res string, err error)
	EditCats(tc entity.TransCatLimit, client string) (err error)
}

type Bot struct {
	api *tgbotapi.BotAPI
	accountant
}

type userChat struct {
	chatID    int64
	messageID int
	userName  string
}

func New(api *tgbotapi.BotAPI, acc accountant) *Bot {
	return &Bot{
		api:        api,
		accountant: acc,
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

func (b *Bot) requestCats(ctx context.Context, page int, query *tgbotapi.CallbackQuery, update *tgbotapi.Update) {
	var (
		userName  string
		chatID    int64
		messageID int
		limit     string
		direction string
	)

	if update == nil {
		userName = query.From.UserName
		chatID = query.Message.Chat.ID
		messageID = query.Message.MessageID
	} else {
		userName = update.SentFrom().UserName
		chatID = update.Message.Chat.ID
		messageID = update.Message.MessageID
	}

	cats, err := b.accountant.GetCatsLimit(ctx, userName, "balance")
	if err != nil {
		return
	}
	options := make([][]string, len(cats))
	for i, v := range cats {
		switch v.Direction.Int16 {
		case -1:
			direction = EMOJI_DEBIT
		case 0:
			direction = EMOJI_DEPOSIT
		case 1:
			direction = EMOJI_CREDIT
		}
		if v.Limit.Valid {
			limit = fmt.Sprintf(" (%d)", v.Limit.Int64)
		} else {
			limit = ""
		}
		options[i] = []string{
			v.Category.String + " " + direction + limit,
			PREFIX_CATEGORY + ":" + v.Category.String,
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

	if update == nil {
		msg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, *resMrkp)
		b.api.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, update.Message.Text+"₽")
		msg.ReplyMarkup = resMrkp
		b.api.Send(msg)
	}
}

func (b *Bot) processNumber(ctx context.Context, u *tgbotapi.Update) (err error) {
	b.requestCats(ctx, 0, nil, u)
	b.deleteMsg(u.Message.Chat.ID, u.Message.MessageID)

	return err
}

func (b *Bot) confirmRecord(query *tgbotapi.CallbackQuery) {
	var (
		amnt, direction int
		cat, descr      string
	)
	amnt, _ = strconv.Atoi(strings.Split(query.Message.Text, "₽")[0])
	direction = 0
	scntSplit := strings.Split(query.Message.Text, "\n")
	cat = strings.Split(scntSplit[0], " на ")[1]
	if len(scntSplit) > 1 {
		descr, _ = strings.CutPrefix(scntSplit[1], EMOJI_COMMENT)
	}

	b.accountant.PostDoc(context.Background(), cat, amnt, descr, fmt.Sprint(query.Message.MessageID), direction, query.From.UserName)
	b.clearMsgReplyMarkup(query.Message.Chat.ID, query.Message.MessageID)
}

func (b *Bot) deleteRecord(query *tgbotapi.CallbackQuery) {
	b.accountant.DeleteDoc(fmt.Sprint(query.Message.MessageID), query.From.UserName)
	b.deleteMsg(query.Message.Chat.ID, query.Message.MessageID)
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

func (b *Bot) requestReply(query *tgbotapi.CallbackQuery, respCode string) {
	b.clearMsgReplyMarkup(query.Message.Chat.ID, query.Message.MessageID)
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, EMOJI_COMMENT+"...")
	msg.ReplyMarkup = getReply()
	b.api.Send(msg)
	waitUserResponeStart(query.From.UserName, respCode, *query.Message)
}

func (b *Bot) requestSubCats(ctx context.Context, page int, query *tgbotapi.CallbackQuery) {
	var cat string
	if len(strings.Split(query.Message.Text, " ")) < 3 {
		cat, _ = strings.CutPrefix(query.Data, PREFIX_CATEGORY+":")
	} else {
		cat = strings.Join(strings.Split(query.Message.Text, " ")[2:], " ")
	}

	subCats, err := b.accountant.GetSubCats(ctx, query.From.UserName, cat)
	fmt.Printf("%#v %d\n %v", subCats, len(subCats), err)
	if len(subCats) == 0 {
		subCats = []string{" "}
	}
	if err != nil {
		return
	}
	options := make([][]string, 0, len(subCats))
	for _, v := range subCats {
		options = append(options, []string{v, PREFIX_SUBCATEGORY + ":" + v})
	}

	mrkp := newKeyboardForm()
	mrkp.setOptions(options)
	mrkp.addNavigationControl(page, nil, []string{EMOJI_KEYBOARD, PREFIX_SUBCATEGORY + ":writeCustom"})
	resMrkp, err := mrkp.getMarkup()
	if err != nil {
		fmt.Println(err)
		return
	}

	msg := tgbotapi.NewEditMessageReplyMarkup(query.Message.Chat.ID, query.Message.MessageID, *resMrkp)
	b.api.Send(msg)
}

// func processReply(u *tgbotapi.Update) (err error) {
// 	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "")
// 	msg.ReplyMarkup = getReply()

// 	gBot.Send(msg)

// 	return
// }

func (b *Bot) checkUser(ctx context.Context, userName string) bool {
	_, f := BotUsers[userName]
	if !f {
		active, err := b.accountant.GetUserStatus(ctx, userName)
		if err != nil {
			log.Println(err)
		}
		if active {
			NewBotUser(userName)
		}
		return active
	}
	return f
}

func (b *Bot) handleUpdate(ctx context.Context, update *tgbotapi.Update) {

	if !b.checkUser(ctx, update.SentFrom().UserName) {
		return
	}

	if update.CallbackQuery != nil {
		b.callbackQueryHandler(ctx, update.CallbackQuery)
		return
	}

	if update.Message != nil {
		if responseIsAwaited(update.SentFrom().UserName) {
			b.responseHandler(update)
			return
		}

		// if update.Message.ReplyToMessage != nil {
		// 	processReply(&update)
		// }

		if update.Message.IsCommand() {
			b.processCommand(ctx, update)
			return
		}

		if len(update.Message.Text) > 0 {
			if _, err := strconv.Atoi(update.Message.Text); err == nil {
				b.processNumber(ctx, update)
				return
			}
		}
	}
}

func (b *Bot) Run(ctx context.Context) (err error) {
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
}
