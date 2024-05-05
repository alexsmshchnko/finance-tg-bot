package tg_bot

import (
	"fmt"
	"strconv"
	"strings"
)

type finMessage struct {
	amount      int
	curr        string
	category    string
	description string
}

func (f *finMessage) parseFinMsg(text string) (fn *finMessage, err error) {
	spltLines := strings.Split(text, "\n")

	spltLine1 := strings.Split(spltLines[0], " на ")
	amntStr, found := strings.CutSuffix(spltLine1[0], f.curr)
	if !found {
		return
	}
	if f.amount, err = strconv.Atoi(amntStr); err != nil {
		return
	}
	fn = f

	if len(spltLine1) < 2 {
		return
	}
	f.category = spltLine1[1]

	if len(spltLines) < 2 {
		return
	}
	f.description, _ = strings.CutPrefix(spltLines[1], EMOJI_COMMENT)

	return
}

func NewFinMsg() *finMessage {
	return &finMessage{curr: "₽"}
}

func (f *finMessage) String() string {
	if f.description != "" {
		return fmt.Sprintf("%d%s на %s\n%s%s", f.amount, f.curr, f.category, EMOJI_COMMENT, f.description)
	} else if f.category != "" {
		return fmt.Sprintf("%d%s на %s", f.amount, f.curr, f.category)
	}
	return fmt.Sprintf("%d%s", f.amount, f.curr)
}

func (f *finMessage) SetAmount(amnt int) {
	f.amount = amnt
}

func (f *finMessage) SetCurr(text string) {
	f.curr = text
}

func (f *finMessage) SetCategory(text string) {
	f.category = text
}

func (f *finMessage) SetDescription(text string) {
	f.description = text
}
