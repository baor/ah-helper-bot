package storage

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/firestore"
	fs "cloud.google.com/go/firestore"

	"github.com/baor/ah-helper-bot/domain"
)

type firestoreAdapter struct {
	client  *fs.Client
	context context.Context
}

const subscriptionCollection = "subscriptions"

// NewFirestoreAdapter creates new adapter
func NewFirestoreAdapter(projectID string) DataStorer {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	adapater := firestoreAdapter{
		client:  client,
		context: ctx,
	}

	log.Printf("Firestore clientto projec '%s' is created", projectID)
	return &adapater
}

func (a *firestoreAdapter) AddSubscription(sub domain.Subscription) {
	_, _, err := a.client.Collection(subscriptionCollection).Add(a.context, sub)
	if err != nil {
		log.Fatalf("Error %v on adding subscription: %v+", err, sub)
	}
}

func (a *firestoreAdapter) GetSubscriptions() []domain.Subscription {
	iter := a.client.Collection(subscriptionCollection).Documents(a.context)
	subs := []domain.Subscription{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		fmt.Println(doc.Data())

		var sub domain.Subscription
		if err := doc.DataTo(&sub); err != nil {
			log.Fatalf("Error when reading data from storage: %v", err)
			return subs
		}
		subs = append(subs, sub)
	}
	return subs
}