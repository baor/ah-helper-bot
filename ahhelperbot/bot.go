package ahhelperbot

import (
	"log"
	"regexp"

	"github.com/baor/ah-helper-bot/domain"
	"github.com/baor/ah-helper-bot/storage"
	"github.com/baor/ah-helper-bot/telegram"
)

// Bot is a telegram bot which returns events and sends messages
type Bot struct {
	messenger telegram.Messenger

	reAddme    *regexp.Regexp
	reRemoveme *regexp.Regexp

	storage storage.DataStorer

	deliveryProvider DeliveryProvider
}

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type pubSubMessage struct {
	Data []byte `json:"data"`
}

// NewBot returns an instance of Bot which implements Messenger interface
// tlgr - is an low-level abstraction for telegram API
func NewBot(storage storage.DataStorer, deliveryProvider DeliveryProvider) *Bot {
	b := Bot{}

	b.reAddme = regexp.MustCompile(`addme (\d{4}\w{2})`)
	b.reRemoveme = regexp.MustCompile(`removeme (\d{4}\w{2})`)
	b.storage = storage

	b.deliveryProvider = deliveryProvider
	return &b
}

// SetMessenger sets messenger, because messager includes message processing
func (b *Bot) SetMessenger(messenger telegram.Messenger) {
	b.messenger = messenger
}

// CheckDeliveries checks delivery for subscripions
func (b *Bot) CheckDeliveries() {
	for _, subscription := range b.storage.GetSubscriptions() {
		deliverySchedule := b.deliveryProvider.Get(subscription.Postcode)
		b.send(domain.Message{
			ChatID: subscription.ChatID,
			Text:   deliverySchedule.String()})
	}
}

// send message to the telegram chat
func (b *Bot) send(msg domain.Message) {
	if len(msg.Text) > 4096 {
		log.Printf("trim too long string %s", msg.Text)
		msg.Text = msg.Text[:4090] + "..."
	}
	b.messenger.Send(msg)
}

func (b *Bot) sendMessageHelp(chatID int64) {
	msg := "Help for the chatbot. Please enter your postcode in format \"addme 1111AA\""
	b.messenger.Send(domain.Message{ChatID: chatID, Text: msg})
}

// DefaultMessageProcessor is a processor for messages to bot
func (b *Bot) DefaultMessageProcessor(msg domain.Message) {
	match := b.reAddme.FindStringSubmatch(msg.Text)
	if match != nil {
		postcode := match[1]
		sub := domain.Subscription{
			ChatID:   msg.ChatID,
			Postcode: postcode,
		}
		log.Printf("message processor add subscription: %+v", sub)
		b.storage.AddSubscription(sub)
		return
	}

	match = b.reRemoveme.FindStringSubmatch(msg.Text)
	if match != nil {
		postcode := match[1]
		sub := domain.Subscription{
			ChatID:   msg.ChatID,
			Postcode: postcode,
		}
		log.Printf("message processor remove subscription: %+v", sub)
		b.storage.RemoveSubscription(sub)
		return
	}

	b.sendMessageHelp(msg.ChatID)
}
