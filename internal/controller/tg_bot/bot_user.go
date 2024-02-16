package tg_bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotUser struct {
	//UserName     string `json:"username,omitempty"`
	//FirstLogin   time.Time
	// diskToken    string
	ResponseWait bool
	//ResponseMsgID int
	ResponseMsg   tgbotapi.Message
	ResponseCode  string
	FinCategories []string
}

var BotUsers map[string]BotUser

// func (b *BotUser) setUserDiskToken(s string) {
// 	b.diskToken = s
// }

// func (b *BotUser) getUserDiskToken() string {
// 	return b.diskToken
// }

func NewBotUser(userName string) {
	BotUsers = map[string]BotUser{userName: BotUser{}}
	log.Printf("BotUsers: %v\n", BotUsers)
}

func responseIsAwaited(userName string) bool {
	return BotUsers[userName].ResponseWait
}

func waitUserResponeStart(userName, respCode string, message tgbotapi.Message) {
	if _, ok := BotUsers[userName]; ok {
		BotUsers[userName] = BotUser{
			ResponseWait: true,
			ResponseMsg:  message,
			ResponseCode: respCode,
		}
	}
	log.Printf("ResponseWaitStart: %v\n", BotUsers[userName])
}

func waitUserResponseComplete(userName string) {
	if _, ok := BotUsers[userName]; ok {
		BotUsers[userName] = BotUser{
			ResponseWait: false,
		}
	}
	log.Printf("ResponseWaitStop: %v\n", BotUsers[userName])
}
