package token

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Payload struct {
	ID        primitive.ObjectID
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewPayload(username string, id primitive.ObjectID) *Payload {
	return &Payload{
		ID:        id,
		Username:  username,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

func (payload *Payload) Valid() error {
	if payload.ExpiresAt.After(payload.CreatedAt) {
		return nil
	}
	return ErrExpiredToken
}
