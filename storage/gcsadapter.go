package storage

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"

	gcs "cloud.google.com/go/storage"

	"github.com/baor/ah-helper-bot/domain"
)

type GcsAdapter struct {
	dbObjectHandle *gcs.ObjectHandle
}

const dbFileName = "ah-helper-db.json"

func NewGcsAdapter(bucketName string) DataStorer {
	s := GcsAdapter{}
	ctx := context.Background()
	client, err := gcs.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	s.dbObjectHandle = client.Bucket(bucketName).Object(dbFileName)

	// check that DB file exists
	_, err = s.dbObjectHandle.NewReader(ctx)
	if err != nil {
		// create empty DB file
		s.writeToDb([]byte("[]"))
	}

	log.Printf("Bucket '%s' with db file '%s' is created", bucketName, dbFileName)
	return &s
}

func (s *GcsAdapter) AddSubscription(sub domain.Subscription) {
	subs := s.readFromDb()
	// TODO: add dedupe check
	subs = append(subs, sub)
	data, err := json.Marshal(subs)
	if err != nil {
		log.Panic(err)
	}
	s.writeToDb(data)
}

func (s *GcsAdapter) readFromDb() []domain.Subscription {
	ctx := context.Background()

	r, err := s.dbObjectHandle.NewReader(ctx)
	if err != nil {
		log.Panic(err)
	}
	defer r.Close()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		log.Panic(err)
	}

	subs := []domain.Subscription{}
	err = json.Unmarshal(data, &subs)
	if err != nil {
		log.Panic(err)
	}

	return subs
}

func (s *GcsAdapter) writeToDb(data []byte) {
	fw := s.dbObjectHandle.NewWriter(context.Background())

	if _, err := fw.Write(data); err != nil {
		log.Panic(err)
	}

	if err := fw.Close(); err != nil {
		log.Panic(err)
	}
}

func (s *GcsAdapter) GetSubscriptions() []domain.Subscription {
	return s.readFromDb()
}
