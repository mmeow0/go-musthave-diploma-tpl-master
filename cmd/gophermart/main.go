package main

import (
	"github.com/mmeow0/gophermart-bonus/internal/app"
	"log"
)

func main() {
	app, err := app.InitializeApp()
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
