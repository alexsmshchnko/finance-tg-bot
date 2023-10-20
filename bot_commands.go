package main

import (
	"finance-tg-bot/internal"
)

func runSync(userName string) (msg string) {
	msg = "\U0001f64c"
	err := internal.SyncDiskFile(userName)
	if err != nil {
		msg = "\U0001f44e"
	}

	return
}
