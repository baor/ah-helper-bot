package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/baor/ah-helper-bot/ahhelperbot"
	"github.com/baor/ah-helper-bot/domain"
	"github.com/baor/ah-helper-bot/storage"
	"github.com/baor/ah-helper-bot/telegram"
)

func getBotToken() string {
	token := os.Getenv("BOT_TELEGRAM_TOKEN")
	if token == "" {
		log.Panic("Empty BOT_TELEGRAM_TOKEN")
	}

	return token
}

func getProjectID() string {
	gcProjectID := os.Getenv("BOT_PROJECT_ID")
	if len(gcProjectID) == 0 {
		log.Panic("Empty BOT_PROJECT_ID")
	}
	log.Printf("BOT_PROJECT_ID: %s", gcProjectID)

	return gcProjectID
}

func main() {
	s := storage.NewFirestoreAdapter(getProjectID())
	eventsCh := make(chan domain.Event, 1)
	bot := ahhelperbot.NewBot(eventsCh, s, &ahhelperbot.DefaultDeliveryProvider{})
	telegramMessenger := telegram.NewMessenger(getBotToken(), bot.DefaultMessageProcessor, 5*time.Second)
	bot.SetMessenger(telegramMessenger)

	http.HandleFunc("/check_deliveries", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("check_deliveries request is received")
		bot.CheckDeliveries()
		log.Printf("check_deliveries is done")
		fmt.Fprint(w, "check_deliveries is done")
	})
	http.ListenAndServe(":8080", nil)
}
