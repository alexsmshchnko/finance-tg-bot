package main

import (
	"finance-tg-bot/internal"
	"log"
)

func runSync(userName string) (msg string) {
	msg = "\U0001f44d"
	err := internal.SyncDiskFile(userName)
	if err != nil {
		log.Println(err)
		msg = "\U0001f44e"
	}

	return
}
