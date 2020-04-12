package telegram

import (
	domain "github.com/baor/ah-helper-bot/ahhelperbot"
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
