package telegram

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/baor/ah-helper-bot/domain"
	botApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type message struct {
	text   string
	chatID int64
}
type Bot struct {
	botAPI         *botApi.BotAPI
	updates        botApi.UpdatesChannel
	messagesToSend chan message
}

// NewBot returns an instance of Bot which implements Messenger interface
func NewBot(token string) Messenger {
	var err error
	b := Bot{}
	b.botAPI, err = botApi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	b.botAPI.Debug = false
	log.Printf("Telegram bot authorized on account %s", b.botAPI.Self.UserName)

	go b.messageSenderLoop()

	return &b
}

func (b *Bot) NewMessageToChat(chatID int64, text string) {
	b.messagesToSend <- message{
		chatID: chatID,
		text:   text,
	}
}

func (b *Bot) GetSubscriptionEvents() <-chan SubscriptionEvent {
	return nil
}

func (b *Bot) UpdatesListener(w http.ResponseWriter, r *http.Request) *domain.Subscription {
	bytes, _ := ioutil.ReadAll(r.Body)

	var update botApi.Update
	json.Unmarshal(bytes, &update)

	return b.parseSubscription(update)
}

// const webhookPath = "ah-helper-webhook"

// func (b *Bot) SetWebHook(selfHost string) {
// 	_, err := b.botAPI.SetWebhook(botApi.NewWebhook(fmt.Sprintf("%s/%s", selfHost, webhookPath)))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

func (b *Bot) messageSenderLoop() {
	for {
		select {
		case msg := <-b.messagesToSend:
			b.sendMessage(msg)
		default:
			time.Sleep(5 * time.Second)
		}
	}
}

func (b *Bot) messageReaderLoop() {
	for {
		select {
		case msg := <-b.updates:
			b.parseSubscription(msg)
		default:
			time.Sleep(1 * time.Second)
		}
	}

	// 	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	// 	msg.ReplyToMessageID = update.Message.MessageID

	// 	bot.Send(msg)
	// }
}

func (b *Bot) parseSubscription(u botApi.Update) *domain.Subscription {
	re := regexp.MustCompile(`add (\d{4}\w{2})`)
	match := re.FindStringSubmatch(u.Message.Text)

	if match == nil {
		msg := "Please enter your postcode in format \"add 1111AA\""
		botMsg := botApi.NewMessage(u.Message.Chat.ID, msg)
		_, err := b.botAPI.Send(botMsg)
		if err != nil {
			log.Panic(err)
		}
		return nil
	}

	postcode := match[1]
	msg := fmt.Sprintf("ChatID: %d, Postcode: %s", u.Message.Chat.ID, postcode)
	botMsg := botApi.NewMessage(u.Message.Chat.ID, msg)
	_, err := b.botAPI.Send(botMsg)
	if err != nil {
		log.Panic(err)
	}
	return &domain.Subscription{
		ChatID:   u.Message.Chat.ID,
		Postcode: postcode,
	}
}

func (b *Bot) sendMessage(msg message) {
	if len(msg.text) > 4096 {
		log.Printf("trim too long string %s", msg.text)
		msg.text = msg.text[:4090] + "..."
	}
	botMsg := botApi.NewMessage(msg.chatID, msg.text)
	botMsg.ParseMode = "HTML"
	log.Println(botMsg)
	_, err := b.botAPI.Send(botMsg)
	if err != nil {
		log.Panic(err)
	}
}
