package main

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
		tgbotapi.NewInlineKeyboardButtonData("Удалить "+EMOJI_CROSS, "OPT:deleteRecord"),
		tgbotapi.NewInlineKeyboardButtonData("Описание "+EMOJI_COMMENT, "OPT:addDescription"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Сохранить "+EMOJI_SAVE, "OPT:saveRecord"),
	),
)

func getCatPageInlineKeyboard(slc []string, page int) *tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	maxPageLen := 5
	pageCnt := len(slc) / maxPageLen

	// log.Println(slc)
	// log.Println(len(slc))
	// log.Println(cap(slc))
	// log.Printf("page: %d\n", page)

	rowsToShow := slc[page*maxPageLen : (page*maxPageLen + maxPageLen)]

	for i := range rowsToShow {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(rowsToShow[i], "CAT:"+rowsToShow[i])))
	}

	var buttons []tgbotapi.InlineKeyboardButton

	if page == 0 {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(EMOJI_NEXT, fmt.Sprintf("PAGE:next:%d:%d", page, pageCnt)))
	} else if page == pageCnt-1 {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(EMOJI_PREV, fmt.Sprintf("PAGE:prev:%d:%d", page, pageCnt)))
	} else {
		buttons = append(buttons,
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_PREV, fmt.Sprintf("PAGE:prev:%d:%d", page, pageCnt)),
			tgbotapi.NewInlineKeyboardButtonData(EMOJI_NEXT, fmt.Sprintf("PAGE:next:%d:%d", page, pageCnt)),
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
