package tg_bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type keyboardMarkup struct {
	options [][]tgbotapi.InlineKeyboardButton
	control [][]tgbotapi.InlineKeyboardButton
}

func newKeyboardForm() *keyboardMarkup {
	return &keyboardMarkup{}
}

func (k *keyboardMarkup) setOptions(options [][]string) {
	res := make([][]tgbotapi.InlineKeyboardButton, 0, len(options))
	for _, v := range options {
		res = append(res, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(v[0], v[1])))
	}
	k.options = res
}

func (k *keyboardMarkup) setControl(control [][][]string) {
	res := make([][]tgbotapi.InlineKeyboardButton, 0, len(control))
	for _, r := range control {
		buttons := make([]tgbotapi.InlineKeyboardButton, 0, len(r))
		for _, b := range r {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(b[0], b[1]))
		}
		res = append(res, buttons)
	}
	k.control = res
}

func (k *keyboardMarkup) getMarkup() (*tgbotapi.InlineKeyboardMarkup, error) {
	if len(k.options) == 0 && len(k.control) == 0 {
		return nil, errors.New("markup is not set")
	}
	mrkp := tgbotapi.NewInlineKeyboardMarkup(append(k.options, k.control...)...)

	return &mrkp, nil
}

func getNavigationControl() {

}

type EditMessageForceReply struct {
	BaseEdit   tgbotapi.BaseEdit
	ForceReply tgbotapi.ForceReply
	// Caption         string
	// ParseMode       string
	// CaptionEntities []MessageEntity
}

// var (
// 	msgExpandedOptionsInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
// 		tgbotapi.NewInlineKeyboardRow(
// 			tgbotapi.NewInlineKeyboardButtonData(EMOJI_CROSS, PREFIX_OPTION+":deleteRecord"),
// 			tgbotapi.NewInlineKeyboardButtonData(EMOJI_SAVE, PREFIX_OPTION+":saveRecord"),
// 		),
// 		tgbotapi.NewInlineKeyboardRow(
// 			tgbotapi.NewInlineKeyboardButtonData("File", PREFIX_OPTION+":attachFile"),
// 			tgbotapi.NewInlineKeyboardButtonData("Location", PREFIX_OPTION+":attachLocation"),
// 			tgbotapi.NewInlineKeyboardButtonData("TIME", PREFIX_OPTION+":changeDate"),
// 			tgbotapi.NewInlineKeyboardButtonData("₽➡$", PREFIX_OPTION+":changeCurrency"),
// 		),
// 	)
// )

func (b *Bot) handleNavigationCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) {
	var (
		err             error
		list            []string
		centerButtonTag string
	)

	prefix := strings.Split(*query.Message.ReplyMarkup.InlineKeyboard[0][0].CallbackData, ":")[0]

	switch prefix {
	case PREFIX_CATEGORY:
		list = BotUsers[query.From.UserName].FinCategories
		if len(list) < 1 {
			fmt.Println("User category cash is empty")
			list, err = b.accountant.GetCats(ctx, query.From.UserName)
			if err != nil {
				log.Println(err)
			}
		}
	case PREFIX_SUBCATEGORY:
		subCat := strings.Join(strings.Split(query.Message.Text, " ")[2:], " ")
		list, _ = b.accountant.GetSubCats(ctx, query.From.UserName, subCat)
		centerButtonTag = PREFIX_SUBCATEGORY + ":" + EMOJI_KEYBOARD
	case PREFIX_SETCATEGORY:
		list, err = b.accountant.GetCats(ctx, query.From.UserName)
		if err != nil {
			log.Println(err)
		}
		list = addButtonToSlice(list, EMOJI_ADD+" (добавить)")
		fmt.Println(list)
	}

	split := strings.Split(query.Data, ":")
	page, err := strconv.Atoi(split[2])
	if err != nil {
		log.Println(err)
	}
	switch split[1] {
	case "next":
		page++
	case "prev":
		page--
	}

	mrkp := getPagedListInlineKeyboard(list, page, prefix, centerButtonTag)
	msg := tgbotapi.NewEditMessageReplyMarkup(query.Message.Chat.ID, query.Message.MessageID, *mrkp)
	b.api.Send(msg)
}

func addButtonToSlice(slc []string, buttonText string) []string {
	slc = append(slc, buttonText)
	return slc
}

func getPagedListInlineKeyboard(slc []string, page int, prefix, centerButtonTag string) *tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var sliceBegin, sliceEnd int

	sliceBegin = page * maxPageLen
	if len(slc)-page*maxPageLen < maxPageLen {
		sliceEnd = len(slc)
	} else {
		sliceEnd = page*maxPageLen + maxPageLen
	}

	rowsToShow := slc[sliceBegin:sliceEnd]

	for i := range rowsToShow {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(rowsToShow[i], prefix+":"+rowsToShow[i])))
	}

	var (
		buttons      []tgbotapi.InlineKeyboardButton
		centerButton tgbotapi.InlineKeyboardButton
	)

	pageCnt := len(slc) / maxPageLen
	if pageCnt*maxPageLen < len(slc) {
		pageCnt++
	}

	if len(centerButtonTag) > 0 {
		centerButton = tgbotapi.NewInlineKeyboardButtonData(EMOJI_KEYBOARD, centerButtonTag)
	}

	if page == 0 && len(slc) > maxPageLen {
		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(" ", " "))
		if len(centerButtonTag) > 0 {
			buttons = append(buttons, centerButton)
		}
		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_NEXT, fmt.Sprintf(PREFIX_PAGE+":next:%d:%d", page, pageCnt)))
	} else if page == pageCnt-1 && page != 0 {
		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_PREV, fmt.Sprintf(PREFIX_PAGE+":prev:%d:%d", page, pageCnt)))
		if len(centerButtonTag) > 0 {
			buttons = append(buttons, centerButton)
		}
		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(" ", " "))
	} else if page > 0 && page < pageCnt-1 {
		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_PREV, fmt.Sprintf(PREFIX_PAGE+":prev:%d:%d", page, pageCnt)))
		if len(centerButtonTag) > 0 {
			buttons = append(buttons, centerButton)
		}
		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_NEXT, fmt.Sprintf(PREFIX_PAGE+":next:%d:%d", page, pageCnt)))
	} else if len(centerButtonTag) > 0 {
		buttons = append(buttons, centerButton)
	}

	if len(buttons) > 0 {
		rows = append(rows, buttons)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return &keyboard
}

func getMsgOptionsKeyboard() *tgbotapi.InlineKeyboardMarkup {
	mrkp := newKeyboardForm()
	mrkp.setControl([][][]string{
		{
			{EMOJI_CROSS, PREFIX_OPTION + ":deleteRecord"},
			{EMOJI_DOWN, PREFIX_OPTION + ":expandOptions"},
			{EMOJI_SAVE, PREFIX_OPTION + ":saveRecord"},
		},
	})
	res, err := mrkp.getMarkup()
	if err != nil {
		return nil
	}
	return res
}

func getReportKeyboard() *tgbotapi.InlineKeyboardMarkup {
	mrkp := newKeyboardForm()
	mrkp.setOptions([][]string{
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

func getSettingsKeyboard() *tgbotapi.InlineKeyboardMarkup {
	mrkp := newKeyboardForm()
	mrkp.setOptions([][]string{
		{"Редактировать категории", PREFIX_SETTING + ":editCategory"},
	})
	mrkp.setControl([][][]string{
		{{EMOJI_CROSS, PREFIX_SETTING + ":cancelSettings"}},
	})
	res, err := mrkp.getMarkup()
	if err != nil {
		return nil
	}
	return res
}

func getReply() *tgbotapi.ForceReply {
	return &tgbotapi.ForceReply{
		ForceReply:            true,
		InputFieldPlaceholder: "Описание",
	}
}
