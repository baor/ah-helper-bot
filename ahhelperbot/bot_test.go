package ahhelperbot

import (
	"fmt"
	"testing"

	"github.com/baor/ah-helper-bot/domain"
	tlg "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/assert"
)

type fakeMessenger struct {
	tlgBotAPI    *tlg.BotAPI
	updatesCh    chan tlg.Update
	sentMessages map[int64]string
}

func newFakeMessenger() *fakeMessenger {
	b := fakeMessenger{}
	b.sentMessages = map[int64]string{}

	b.updatesCh = make(chan tlg.Update, 1)
	return &b
}

func (b *fakeMessenger) Send(m domain.Message) {
	b.sentMessages[m.ChatID] = m.Text
}

type fakeDeliveryProvider struct {
	date string
}

func (p *fakeDeliveryProvider) Get(postcode string) DeliverySchedule {
	resp := DeliverySchedule{}
	resp[p.date] = []DeliveryTimeSlotBase{
		{
			From: postcode,
		},
	}
	return resp
}

type fakeDataStorer struct {
	subscriptions map[int64]domain.Subscription
}

func (s *fakeDataStorer) AddSubscription(subscription domain.Subscription) {
	if s.subscriptions == nil {
		s.subscriptions = make(map[int64]domain.Subscription)
	}
	s.subscriptions[subscription.ChatID] = subscription
}

func (s *fakeDataStorer) RemoveSubscription(subscription domain.Subscription) {
	delete(s.subscriptions, subscription.ChatID)
}

func (s *fakeDataStorer) GetSubscriptions() []domain.Subscription {
	subs := []domain.Subscription{}
	for _, v := range s.subscriptions {
		subs = append(subs, v)
	}
	return subs
}

func TestBot_SendMessage(t *testing.T) {
	fakeMessenger := newFakeMessenger()
	storage := fakeDataStorer{}
	bot := NewBot(&storage, &fakeDeliveryProvider{})
	bot.SetMessenger(fakeMessenger)

	// Act
	bot.send(domain.Message{ChatID: 1, Text: "test"})

	assert.Equal(t, "test", fakeMessenger.sentMessages[1])
}

func TestBotMessageProcessor_ProcessHelp(t *testing.T) {
	fakeMessenger := newFakeMessenger()
	storage := fakeDataStorer{}
	bot := NewBot(&storage, &fakeDeliveryProvider{})
	bot.SetMessenger(fakeMessenger)

	msg := domain.Message{
		ChatID: 1,
		Text:   "help",
	}
	// Act
	bot.DefaultMessageProcessor(msg)

	sentMsg := fakeMessenger.sentMessages[1]
	assert.Contains(t, sentMsg, "Help")
}

func TestBotMessageProcessor_ProcessAdd(t *testing.T) {
	fakeMessenger := newFakeMessenger()
	storage := fakeDataStorer{}
	bot := NewBot(&storage, &fakeDeliveryProvider{})
	bot.SetMessenger(fakeMessenger)

	msg := domain.Message{
		ChatID: 1,
		Text:   "addme 1234AA",
	}
	// Act
	bot.DefaultMessageProcessor(msg)

	assert.Equal(t, domain.Subscription{
		ChatID:   1,
		Postcode: "1234AA",
	}, storage.subscriptions[1])
}

func TestBotMessageProcessor_ProcessRemove(t *testing.T) {
	fakeMessenger := newFakeMessenger()
	storage := fakeDataStorer{
		subscriptions: map[int64]domain.Subscription{
			1: domain.Subscription{
				ChatID:   1,
				Postcode: "1234AA",
			},
		},
	}
	bot := NewBot(&storage, &fakeDeliveryProvider{})
	bot.SetMessenger(fakeMessenger)

	msg := domain.Message{
		ChatID: 1,
		Text:   "removeme 1234AA",
	}

	// Act
	bot.DefaultMessageProcessor(msg)

	assert.Equal(t, 0, len(storage.subscriptions))
}

func TestBotDelivery_Get(t *testing.T) {
	fakeMessenger := newFakeMessenger()
	postcode := "1234AA"
	storage := fakeDataStorer{
		subscriptions: map[int64]domain.Subscription{
			1: domain.Subscription{
				ChatID:   1,
				Postcode: "1234AA",
			},
		},
	}

	provider := fakeDeliveryProvider{
		date: "01-01-1970",
	}
	bot := NewBot(&storage, &provider)
	bot.SetMessenger(fakeMessenger)

	// Act
	bot.CheckDeliveries()

	sentMsg := fakeMessenger.sentMessages[1]
	assert.Contains(t, sentMsg, fmt.Sprintf("%s:%s-", provider.date, postcode))
}
