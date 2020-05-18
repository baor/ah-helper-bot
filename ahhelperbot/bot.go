package ahhelperbot

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/baor/ah-helper-bot/domain"
	"github.com/baor/ah-helper-bot/storage"
	"github.com/baor/ah-helper-bot/telegram"
)

// Bot is a telegram bot which returns events and sends messages
type Bot struct {
	messenger telegram.Messenger

	reAddme         *regexp.Regexp
	reRemoveme      *regexp.Regexp
	reCheckDelivery *regexp.Regexp

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
		b.checkDelivery(subscription)
	}
}

func (b *Bot) checkDeliveryByID(c domain.ChatID) {
	sub := b.storage.GetSubscriptionByID(c)
	b.checkDelivery(sub)
}

func (b *Bot) checkDelivery(subscription domain.Subscription) {
	if subscription.ChatID == 0 {
		return
	}

	if len(subscription.Postcode) == 0 {
		b.send(domain.Message{
			ChatID: subscription.ChatID,
			Text:   "Postcode was not found. Try to register again with /addme 1234AB"})
		return
	}

	deliverySchedule := b.deliveryProvider.Get(subscription.Postcode)
	scheduleText := deliverySchedule.String()
	if len(scheduleText) == 0 {
		scheduleText = fmt.Sprintf("No deliveries available for %s", subscription.Postcode)
	}
	b.send(domain.Message{
		ChatID: subscription.ChatID,
		Text:   scheduleText})
}

// send message to the telegram chat
func (b *Bot) send(msg domain.Message) {
	if len(msg.Text) > 4096 {
		log.Printf("trim too long string %s", msg.Text)
		msg.Text = msg.Text[:4090] + "..."
	}
	b.messenger.Send(msg)
}

func (b *Bot) sendMessageHelp(chatID domain.ChatID) {
	msg := `Help for the AH chatbot.
	+ In order to register or update information, please enter your postcode in format 
	/addme 1234AB

	+ To remove your registration, enter
	/unsubscribe

	+ To check available deliveries for your postcode enter
	/check

	any other input will show this message
	`
	b.messenger.Send(domain.Message{ChatID: chatID, Text: msg})
}

// DefaultMessageProcessor is a processor for messages to bot
func (b *Bot) DefaultMessageProcessor(msg domain.Message) {
	if strings.HasPrefix(msg.Text, "/check") {
		b.checkDeliveryByID(msg.ChatID)
		return
	}

	if strings.HasPrefix(msg.Text, "/unsubscribe") {
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

	b.sendMessageHelp(msg.ChatID)
}
