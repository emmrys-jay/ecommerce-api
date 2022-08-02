package auth

import (
	"time"
)

type Payload struct {
	ID        string
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewPayload(username string, id string) *Payload {
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
