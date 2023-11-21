package main

import (
	"context"
	"log"
)

func runSync(userName string) (msg string) {
	msg = "\U0001f44d"
	err := sync.MigrateFromCloud(context.Background(), userName)
	if err != nil {
		log.Println(err)
		msg = "\U0001f44e"
	}

	return
}
