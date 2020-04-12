package storage

import (
	domain "github.com/baor/ah-helper-bot/ahhelperbot"
)

// DataStorer to store chats and postcodes
type DataStorer interface {
	AddSubscription(domain.Subscription)
	GetSubscriptions() []domain.Subscription
}
