package tg_bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type EditMessageForceReply struct {
	BaseEdit   tgbotapi.BaseEdit
	ForceReply tgbotapi.ForceReply
	// Caption         string
	// ParseMode       string
	// CaptionEntities []MessageEntity
}

var (
	msgOptionsInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_CROSS, PREFIX_OPTION+":deleteRecord"),
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_DOWN, PREFIX_OPTION+":expandOptions"),
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_SAVE, PREFIX_OPTION+":saveRecord"),
		),
	)
	msgExpandedOptionsInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_CROSS, PREFIX_OPTION+":deleteRecord"),
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_SAVE, PREFIX_OPTION+":saveRecord"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("File", PREFIX_OPTION+":attachFile"),
			tgbotapi.NewInlineKeyboardButtonData("Location", PREFIX_OPTION+":attachLocation"),
			tgbotapi.NewInlineKeyboardButtonData("TIME", PREFIX_OPTION+":changeDate"),
			tgbotapi.NewInlineKeyboardButtonData("₽➡$", PREFIX_OPTION+":changeCurrency"),
		),
	)
	msgReportTypeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Текущий месяц", PREFIX_REPORT+":monthReport:current"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Предыдущий месяц", PREFIX_REPORT+":monthReport:previous"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_CROSS, PREFIX_REPORT+":cancelReport"),
		),
	)
	msgSettingsTypeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Редактировать категории", PREFIX_SETTING+":editCategory"),
		),
	)
)

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
	return &msgOptionsInlineKeyboard
}

func getReportKeyboard() *tgbotapi.InlineKeyboardMarkup {
	return &msgReportTypeKeyboard
}

func getSettingsKeyboard() *tgbotapi.InlineKeyboardMarkup {
	return &msgSettingsTypeKeyboard
}

func getReply() *tgbotapi.ForceReply {
	return &tgbotapi.ForceReply{
		ForceReply:            true,
		InputFieldPlaceholder: "Описание",
	}
}
