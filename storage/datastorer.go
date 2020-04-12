package storage

import (
	"github.com/baor/ah-helper-bot/domain"
)

// DataStorer to store chats and postcodes
type DataStorer interface {
	AddSubscription(domain.Subscription)
	GetSubscriptions() []domain.Subscription
}
