package ahhelperbot

import (
	"github.com/baor/ah-helper-bot/storage"
	"github.com/baor/ah-helper-bot/telegram"
)

type context struct {
	tlg telegram.Messenger
	st  storage.DataStorer
}

// func updateLoop(c context) {
// 	log.Printf("Update feed with delay %v", c.delay)
// 	for {
// 		updateFeedToChannel(c)
// 		time.Sleep(c.delay)c
// 	}
// }
