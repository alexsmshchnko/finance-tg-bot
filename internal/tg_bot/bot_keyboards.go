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

var msgOptionsInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Удалить "+EMOJI_CROSS, PREFIX_OPTION+":deleteRecord"),
		tgbotapi.NewInlineKeyboardButtonData("Описание "+EMOJI_COMMENT, PREFIX_OPTION+":addDescription"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Сохранить "+EMOJI_SAVE, PREFIX_OPTION+":saveRecord"),
	),
)

func getPagedListInlineKeyboard(slc []string, page int, prefix string) *tgbotapi.InlineKeyboardMarkup {
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

	var buttons []tgbotapi.InlineKeyboardButton

	pageCnt := len(slc) / maxPageLen
	if pageCnt*maxPageLen < len(slc) {
		pageCnt++
	}

	if page == 0 && len(slc) > maxPageLen {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(EMOJI_NEXT, fmt.Sprintf(PREFIX_PAGE+":next:%d:%d", page, pageCnt)))
	} else if page == pageCnt-1 && page != 0 {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(EMOJI_PREV, fmt.Sprintf(PREFIX_PAGE+":prev:%d:%d", page, pageCnt)))
	} else if page > 0 && page < pageCnt-1 {
		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_PREV, fmt.Sprintf(PREFIX_PAGE+":prev:%d:%d", page, pageCnt)),
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_NEXT, fmt.Sprintf(PREFIX_PAGE+":next:%d:%d", page, pageCnt)),
		)
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

// func getMenuKeyboard() *tgbotapi.ReplyKeyboardMarkup {
// 	return &menuKeyboard
// }

func getReply() *tgbotapi.ForceReply {
	return &tgbotapi.ForceReply{
		ForceReply:            true,
		InputFieldPlaceholder: "Описание",
	}
}
