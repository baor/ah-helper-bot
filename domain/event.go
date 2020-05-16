package domain

// EventType - type of events which can be triggered from telegram chat
type EventType int

// EventType - enums
const (
	EventTypeAdd EventType = iota + 1
	EventTypeRemove
)

// Event describes outgoing event after message processing
type Event struct {
	Type     EventType
	ChatID   int64
	Postcode string
}
