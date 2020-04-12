package telegram

import (
	"fmt"
	"log"
	"regexp"
	"time"

	botApi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type message struct {
	text   string
	chatID int64
}
type Bot struct {
	botAPI             *botApi.BotAPI
	subscriptionEvents chan SubscriptionEvent
	messagesToSend     chan message
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

	u := botApi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.botAPI.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	// Optional: wait for updates and clear them if you don't want to handle
	// a large backlog of old messages
	time.Sleep(time.Millisecond * 500)
	updates.Clear()

	for {
		select {
		case msg := <-updates:
			b.requestResponse(msg)
		default:
			time.Sleep(1 * time.Second)
		}
	}

	// 	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	// 	msg.ReplyToMessageID = update.Message.MessageID

	// 	bot.Send(msg)
	// }
}

func (b *Bot) requestResponse(u botApi.Update) {
	re := regexp.MustCompile(`add (\d{4}\w{2})`)
	match := re.FindStringSubmatch(u.Message.Text)

	msg := ""
	if match == nil {
		msg = "Please enter your postcode in format \"add 1111AA\""
	} else {
		postcode := match[1]
		msg = fmt.Sprintf("ChatID: %d, Postcode: %s", u.Message.Chat.ID, postcode)
	}

	botMsg := botApi.NewMessage(u.Message.Chat.ID, msg)
	_, err := b.botAPI.Send(botMsg)
	if err != nil {
		log.Panic(err)
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

func (b *Bot) GetSubscriptionEvents() <-chan SubscriptionEvent {
	return b.subscriptionEvents
}
