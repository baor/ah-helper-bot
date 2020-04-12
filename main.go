package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/baor/ah-helper-bot/storage"

	"github.com/baor/ah-helper-bot/telegram"
)

func getBotToken() string {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Panic("Empty TELEGRAM_BOT_TOKEN")
	}

	return token
}

func getGcsBucketName() string {
	gcsBucketName := os.Getenv("BOT_GCS_BUCKET_NAME")
	if len(gcsBucketName) == 0 {
		log.Panic("Empty BOT_GCS_BUCKET_NAME")
	}
	log.Printf("BOT_GCS_BUCKET_NAME: %s", gcsBucketName)

	return gcsBucketName
}

// EntrypointUpdates all messages from user are langing here
func EntrypointUpdates(w http.ResponseWriter, r *http.Request) {
	b := telegram.NewBot(getBotToken())
	subscription := b.UpdatesListener(w, r)
	if subscription == nil {
		return
	}
	s := storage.NewGcsAdapter(getGcsBucketName())
	s.AddSubscription(*subscription)
}

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// EntrypointCheck all update pubsub requests are landing here
func EntrypointCheck(ctx context.Context, m PubSubMessage) error {
	log.Println(string(m.Data))
	getDeliveries("2033BA")
	return nil
}

// func EntrypointSetWebHook(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprint(w, html.EscapeString(d.Message))
// }

func main() {
	log.Panic("Main does nothing. Main package is a set of cloud funcs")
}

func getDeliveries(postcode string) {
	client := http.Client{Timeout: 20 * time.Second}

	url := "https://www.ah.nl/service/rest/delegate?url=%2Fkies-moment%2Fbezorgen%2F" + postcode
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("authority", "www.ah.nl")
	req.Header.Add("accept", "application/json")
	req.Header.Add("x-requested-with", "XMLHttpRequest")
	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.162 Safari/537.36")
	req.Header.Add("x-order-mode", "false")
	req.Header.Add("accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("x-requested-with", "XMLHttpRequest")
	req.Header.Add("x-breakpoint", "medium")
	req.Header.Add("sec-fetch-site", "same-origin")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("referer", "https://www.ah.nl/mijnlijst")
	req.Header.Add("accept-language", "en-US,en;q=0.9,nl-NL;q=0.8,nl;q=0.7,ru-RU;q=0.6,ru;q=0.5")

	resp, err := client.Do(req)
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%q", dump)
}
