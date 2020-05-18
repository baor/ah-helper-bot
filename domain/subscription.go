package domain

import "strconv"

// ChatID is a type for chatIDs
type ChatID int64

func (c ChatID) String() string {
	return strconv.FormatInt(int64(c), 10)
}

// Subscription is a datastructure in DB
type Subscription struct {
	ChatID   ChatID
	Postcode string
}
