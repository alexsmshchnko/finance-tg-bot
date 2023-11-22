package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type EditMessageForceReply struct {
	BaseEdit   tgbotapi.BaseEdit
	ForceReply tgbotapi.ForceReply
	// Caption         string
	// ParseMode       string
	// CaptionEntities []MessageEntity
}

// var menuKeyboard = tgbotapi.NewReplyKeyboard(
// 	tgbotapi.NewKeyboardButtonRow(
// 		tgbotapi.NewKeyboardButton("1"),
// 		tgbotapi.NewKeyboardButton("2"),
// 		tgbotapi.NewKeyboardButton("3"),
// 	),
// 	tgbotapi.NewKeyboardButtonRow(
// 		tgbotapi.NewKeyboardButton("4"),
// 		tgbotapi.NewKeyboardButton("5"),
// 		tgbotapi.NewKeyboardButton("6"),
// 	),
// 	tgbotapi.NewKeyboardButtonRow(
// 		tgbotapi.NewKeyboardButton("7"),
// 		tgbotapi.NewKeyboardButton("8"),
// 		tgbotapi.NewKeyboardButton("9"),
// 	),
// )

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

	log.Println(slc)
	log.Println(len(slc))
	log.Println(cap(slc))
	log.Printf("page: %d\n", page)

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

	// if currentPage > 0 {
	// rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(EMOJI_PREV, fmt.Sprintf("PAGE:prev:%d:%d", currentPage, count)),
	// 	tgbotapi.NewInlineKeyboardButtonData(EMOJI_NEXT, fmt.Sprintf("PAGE:next:%d:%d", currentPage, count))))
	// // }

	// if currentPage < maxPages-1 {
	// rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(EMOJI_NEXT, fmt.Sprintf("pager:next:%d:%d", currentPage, count))))
	// }

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return &keyboard
}

func getCatInlineKeyboard(slc []string, page int, pageCnt int) *tgbotapi.InlineKeyboardMarkup {
	// var rows [][]tgbotapi.InlineKeyboardButton

	// for i := range slc {
	// 	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(slc[i], "CAT:"+slc[i])))
	// }

	// keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	// return &keyboard
	return getCatPageInlineKeyboard(slc, page)
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
