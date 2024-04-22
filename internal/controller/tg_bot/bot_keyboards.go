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
	res := make([][]tgbotapi.InlineKeyboardButton, len(options))
	for i, v := range options {
		res[i] = tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(v[0], v[1]))
	}
	k.options = res
}

func (k *keyboardMarkup) setControl(control [][][]string) {
	res := make([][]tgbotapi.InlineKeyboardButton, len(control))
	for i, r := range control {
		buttons := make([]tgbotapi.InlineKeyboardButton, len(r))
		for j, b := range r {
			buttons[j] = tgbotapi.NewInlineKeyboardButtonData(b[0], b[1])
		}
		res[i] = buttons
	}
	k.control = res
}

func (k *keyboardMarkup) addNavigationControl(page int, firstLeftButton, centerButton []string) {
	var sliceBegin, sliceEnd, pageCnt, slcFullLen int

	//cut options to show
	sliceBegin = page * maxPageLen
	slcFullLen = len(k.options)
	if slcFullLen-page*maxPageLen < maxPageLen {
		sliceEnd = slcFullLen
	} else {
		sliceEnd = page*maxPageLen + maxPageLen
	}
	k.options = k.options[sliceBegin:sliceEnd]

	//add navigation
	pageCnt = slcFullLen / maxPageLen
	if pageCnt*maxPageLen < slcFullLen {
		pageCnt++
	}

	navControl := make([][]string, 0, 3)

	if page == 0 && slcFullLen > maxPageLen {
		if len(firstLeftButton) > 0 {
			navControl = append(navControl, []string{firstLeftButton[0], firstLeftButton[1]})
		} else {
			navControl = append(navControl, []string{" ", " "})
		}
		if len(centerButton) > 0 {
			navControl = append(navControl, []string{centerButton[0], centerButton[1]})
		}
		navControl = append(navControl, []string{EMOJI_NEXT, fmt.Sprintf(PREFIX_PAGE+":next:%d:%d", page, pageCnt)})

	} else if page == pageCnt-1 && page != 0 {
		navControl = append(navControl, []string{EMOJI_PREV, fmt.Sprintf(PREFIX_PAGE+":prev:%d:%d", page, pageCnt)})
		if len(centerButton) > 0 {
			navControl = append(navControl, []string{centerButton[0], centerButton[1]})
		}
		navControl = append(navControl, []string{" ", " "})
	} else if page > 0 && page < pageCnt-1 {
		navControl = append(navControl, []string{EMOJI_PREV, fmt.Sprintf(PREFIX_PAGE+":prev:%d:%d", page, pageCnt)})
		if len(centerButton) > 0 {
			navControl = append(navControl, []string{centerButton[0], centerButton[1]})
		}
		navControl = append(navControl, []string{EMOJI_NEXT, fmt.Sprintf(PREFIX_PAGE+":next:%d:%d", page, pageCnt)})
	} else if len(centerButton) > 0 {
		navControl = append(navControl, []string{centerButton[0], centerButton[1]})
	}

	k.setControl([][][]string{navControl})

}

func (k *keyboardMarkup) getMarkup() (*tgbotapi.InlineKeyboardMarkup, error) {
	if len(k.options) == 0 && len(k.control) == 0 {
		return nil, errors.New("markup is not set")
	}
	mrkp := tgbotapi.NewInlineKeyboardMarkup(append(k.options, k.control...)...)

	return &mrkp, nil
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

func (b *Bot) handleNavigationCallbackQuery(ctx context.Context, q *tgbotapi.CallbackQuery) {
	var (
		page int
		err  error
	)

	prefix := strings.Split(*q.Message.ReplyMarkup.InlineKeyboard[0][0].CallbackData, ":")[0]

	split := strings.Split(q.Data, ":")
	page, err = strconv.Atoi(split[2])
	if err != nil {
		log.Println(err)
	}
	switch split[1] {
	case "next":
		page++
	case "prev":
		page--
	}

	switch prefix {
	case PREFIX_CATEGORY:
		b.requestCats(ctx, page, q, nil)
	case PREFIX_SUBCATEGORY:
		b.requestSubCats(ctx, page, q)
	case PREFIX_SETCATEGORY:
		requestCategoriesKeyboardEditor(b, ctx, page, &userChat{q.Message.Chat.ID, q.Message.MessageID, q.From.UserName})
	}
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

func getMsgExpOptionsKeyboard() *tgbotapi.InlineKeyboardMarkup {
	mrkp := newKeyboardForm()
	mrkp.setOptions([][]string{
		{"Деньги -> время", PREFIX_OPTION + ":money2Time"},
	})
	mrkp.setControl([][][]string{
		{
			{EMOJI_CROSS, PREFIX_OPTION + ":deleteRecord"},
			{EMOJI_SAVE, PREFIX_OPTION + ":saveRecord"},
		},
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
		{"Редактировать категории и лимиты", PREFIX_SETTING + ":editCategory"},
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
