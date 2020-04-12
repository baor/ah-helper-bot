package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/baor/ah-helper-bot/storage"
)

// HandlerStatus handler returns applications status
func status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "bot status is OK")
}

// Entrypoint for all requests
func Entrypoint(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		fmt.Fprint(w, "Hello World!")
		return
	}
	if d.Message == "" {
		fmt.Fprint(w, "Hello World!")
		return
	}
	fmt.Fprint(w, html.EscapeString(d.Message))
}

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Panic("Empty TELEGRAM_BOT_TOKEN")
	}
	//bot := telegram.NewBot(token)

	var s storage.DataStorer

	gcsBucketName := os.Getenv("BOT_GCS_BUCKET_NAME")
	if len(gcsBucketName) > 0 {
		log.Printf("BOT_GCS_BUCKET_NAME: %s", gcsBucketName)
		s = storage.NewGcsAdapter(gcsBucketName)
	}
	//TODO remove
	s.GetSubscriptions()

	delayMinStr := os.Getenv("BOT_DELAY_MIN")
	log.Printf("BOT_DELAY_MIN: %s", delayMinStr)
	delayMinInt64, err := strconv.ParseInt(delayMinStr, 10, 32)
	if err != nil || delayMinInt64 <= 0 {
		delayMinInt64 = 30
	}
	//delay := time.Duration(delayMinInt64) * time.Minute

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	http.HandleFunc("/", status)
	http.ListenAndServe(":"+port, nil)
}
