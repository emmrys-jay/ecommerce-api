package auth

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TokenMaker struct {
	SecretKey []byte
}

var (
	ErrInvalidToken = fmt.Errorf("error: invalid token")
	ErrExpiredToken = fmt.Errorf("error: expired token")
)

func NewTokenMaker() (*TokenMaker, error) {
	secretKey, err := getSecretKey()
	if err != nil {
		return nil, err
	}

	return &TokenMaker{secretKey}, nil
}

func getSecretKey() ([]byte, error) {
	secretKey := os.Getenv("SECRET_KEY")

	if len(secretKey) < 16 {
		return nil, fmt.Errorf("error: validation failed on secret key")
	}
	return []byte(secretKey), nil
}

func (maker *TokenMaker) CreateToken(username string, id primitive.ObjectID) (string, error) {
	payload := NewPayload(username, id)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	tokenString, err := token.SignedString(maker.SecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (maker *TokenMaker) VerifyToken(tokenString string) (*Payload, error) {
	keyfunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return maker.SecretKey, nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &Payload{}, keyfunc)
	if err != nil {
		return nil, ErrInvalidToken
	}

	payload, ok := token.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}
