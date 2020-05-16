package telegram

import (
	"log"
	"time"

	"github.com/baor/ah-helper-bot/domain"
	tlg "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Messenger is an inteface which describes basic messenger functionality
type Messenger interface {
	Send(m domain.Message)
}

// MessageProcessor is a function to process updates in telegram chat
type MessageProcessor func(msg domain.Message)

// tlgMessenger is an adapter for telegram bot functionality
type tlgMessenger struct {
	botAPI           *tlg.BotAPI
	updatesCh        tlg.UpdatesChannel
	messageProcessor MessageProcessor
}

// NewMessenger is a constructor
func NewMessenger(token string, messageProcessor MessageProcessor, delay time.Duration) Messenger {
	var err error

	if token == "" {
		log.Println("Warning! Token is empty! Return nil adapter")
		return nil
	}

	a := tlgMessenger{}
	a.botAPI, err = tlg.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	a.botAPI.Debug = false
	log.Printf("Telegram bot authorized on account %s", a.botAPI.Self.UserName)

	u := tlg.NewUpdate(0)
	u.Timeout = 60

	a.updatesCh, err = a.botAPI.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	a.messageProcessor = messageProcessor

	go a.updatesListener(delay)

	return &a
}

// Send will send a Chattable item to Telegram.
//
// It requires the Chattable to send.
func (a *tlgMessenger) Send(m domain.Message) {
	botMsg := tlg.NewMessage(m.ChatID, m.Text)
	log.Printf("Send message: %v", botMsg)
	_, err := a.botAPI.Send(botMsg)
	if err != nil {
		log.Panic(err)
	}
}

func (a *tlgMessenger) updatesListener(delay time.Duration) {
	for {
		select {
		case u := <-a.updatesCh:
			a.messageProcessor(
				domain.Message{
					ChatID: u.Message.Chat.ID,
					Text:   u.Message.Text,
				},
			)
		default:
			time.Sleep(delay)
		}
	}
}
