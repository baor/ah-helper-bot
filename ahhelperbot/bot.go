package ahhelperbot

import (
	"fmt"
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

	b.reAddme = regexp.MustCompile(`\/addme (\d{4}\w{2})`)
	b.reRemoveme = regexp.MustCompile(`\/unsubscribe`)
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
		if len(subscription.Postcode) == 0 {
			continue
		}
		deliverySchedule := b.deliveryProvider.Get(subscription.Postcode)
		scheduleText := deliverySchedule.String()
		if len(scheduleText) == 0 {
			scheduleText = "delivery not available"
		}
		b.send(domain.Message{
			ChatID: subscription.ChatID,
			Text:   scheduleText})
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
	msg := `Help for the AH chatbot.
	+ In order to register or update information, please enter your postcode in format 
	/addme 1234AB

	+ To remove your registration, enter
	/unsubscribe

	any other input will show this message
	`
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
		b.messenger.Send(domain.Message{
			ChatID: msg.ChatID,
			Text:   fmt.Sprintf("Subscription for postcode %s was successful", postcode),
		})
		return
	}

	if b.reRemoveme.MatchString(msg.Text) {
		sub := domain.Subscription{
			ChatID: msg.ChatID,
		}
		log.Printf("message processor remove subscription: %+v", sub)
		b.storage.RemoveSubscription(sub)
		b.messenger.Send(domain.Message{
			ChatID: msg.ChatID,
			Text:   fmt.Sprintf("Subscription was removed"),
		})
		return
	}

	b.sendMessageHelp(msg.ChatID)
}
