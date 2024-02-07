package main

import (
	"log"

	"finance-tg-bot/config"
	"finance-tg-bot/internal/app"
)

func main() {
	//Configuration
	cfg := config.Get()

	//Run
	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
