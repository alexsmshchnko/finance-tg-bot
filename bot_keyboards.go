package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type EditMessageForceReply struct {
	BaseEdit   tgbotapi.BaseEdit
	ForceReply tgbotapi.ForceReply
	// Caption         string
	// ParseMode       string
	// CaptionEntities []MessageEntity
}

var menuKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("1"),
		tgbotapi.NewKeyboardButton("2"),
		tgbotapi.NewKeyboardButton("3"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("4"),
		tgbotapi.NewKeyboardButton("5"),
		tgbotapi.NewKeyboardButton("6"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("7"),
		tgbotapi.NewKeyboardButton("8"),
		tgbotapi.NewKeyboardButton("9"),
	),
)

var msgOptionsInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Удалить "+EMOJI_CROSS, "OPT:deleteRecord"),
		tgbotapi.NewInlineKeyboardButtonData("Описание "+EMOJI_COMMENT, "OPT:addDescription"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Сохранить "+EMOJI_SAVE, "OPT:saveRecord"),
	),
)

func getCatInlineKeyboard(slc []string) *tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for i := range slc {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(slc[i], "CAT:"+slc[i])))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return &keyboard
}

func getMsgOptionsKeyboard() *tgbotapi.InlineKeyboardMarkup {
	return &msgOptionsInlineKeyboard
}

func getMenuKeyboard() *tgbotapi.ReplyKeyboardMarkup {
	return &menuKeyboard
}

func getReply() *tgbotapi.ForceReply {
	return &tgbotapi.ForceReply{
		ForceReply:            true,
		InputFieldPlaceholder: "Описание",
	}
}
