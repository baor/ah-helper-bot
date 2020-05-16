package ahhelperbot

import (
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
	response string
}

func (p *fakeDeliveryProvider) Get(postcode string) string {
	return postcode + p.response
}

type fakeDataStorer struct {
	subscriptions []domain.Subscription
}

func (s *fakeDataStorer) AddSubscription(subscription domain.Subscription) {
	s.subscriptions = append(s.subscriptions, subscription)

}
func (s *fakeDataStorer) GetSubscriptions() []domain.Subscription {
	return s.subscriptions
}

func TestBot_SendMessage(t *testing.T) {
	fakeMessenger := newFakeMessenger()
	eventsCh := make(chan domain.Event, 1)
	bot := NewBot(eventsCh, &fakeDataStorer{}, &fakeDeliveryProvider{})
	bot.SetMessenger(fakeMessenger)

	// Act
	bot.send(domain.Message{ChatID: 1, Text: "test"})

	assert.Equal(t, "test", fakeMessenger.sentMessages[1])
}

func TestBotMessageProcessor_ProcessHelp(t *testing.T) {
	fakeMessenger := newFakeMessenger()
	eventsCh := make(chan domain.Event, 1)
	bot := NewBot(eventsCh, &fakeDataStorer{}, &fakeDeliveryProvider{})
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
	eventsCh := make(chan domain.Event, 1)
	bot := NewBot(eventsCh, &fakeDataStorer{}, &fakeDeliveryProvider{})
	bot.SetMessenger(fakeMessenger)

	msg := domain.Message{
		ChatID: 1,
		Text:   "addme 1234AA",
	}
	// Act
	bot.DefaultMessageProcessor(msg)

	event := <-eventsCh
	assert.Equal(t, domain.Event{
		Type:     domain.EventTypeAdd,
		ChatID:   1,
		Postcode: "1234AA",
	}, event)
}

func TestBotMessageProcessor_ProcessRemove(t *testing.T) {
	fakeMessenger := newFakeMessenger()
	eventsCh := make(chan domain.Event, 1)
	bot := NewBot(eventsCh, &fakeDataStorer{}, &fakeDeliveryProvider{})
	bot.SetMessenger(fakeMessenger)

	msg := domain.Message{
		ChatID: 1,
		Text:   "removeme 1234AA",
	}

	// Act
	bot.DefaultMessageProcessor(msg)

	event := <-eventsCh
	assert.Equal(t, domain.Event{
		Type:     domain.EventTypeRemove,
		ChatID:   1,
		Postcode: "1234AA",
	}, event)
}

func TestBotDelivery_Get(t *testing.T) {
	fakeMessenger := newFakeMessenger()
	eventsCh := make(chan domain.Event, 1)

	storage := fakeDataStorer{
		subscriptions: []domain.Subscription{
			domain.Subscription{
				ChatID:   1,
				Postcode: "1234AA",
			},
		},
	}

	provider := fakeDeliveryProvider{
		response: "-OK",
	}
	bot := NewBot(eventsCh, &storage, &provider)
	bot.SetMessenger(fakeMessenger)

	// Act
	bot.CheckDeliveries()

	sentMsg := fakeMessenger.sentMessages[1]
	assert.Contains(t, sentMsg, "1234AA-OK")
}
