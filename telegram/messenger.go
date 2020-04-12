package telegram

import (
	"github.com/baor/ah-helper-bot/domain"
)

type Messenger interface {
	NewMessageToChat(chatID int64, text string)
	GetSubscriptionEvents() <-chan SubscriptionEvent
}

type SubscriptionAction int

const (
	Add SubscriptionAction = iota
	Remove
)

type SubscriptionEvent struct {
	Action       SubscriptionAction
	Subscription domain.Subscription
}
