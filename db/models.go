package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `json:"_id" binding:"required"`
	Username string             `json:"username" binding:"required"`
	Password string             `json:"password" binding:"required"`
	Fullname string             `json:"fullname" binding:"required"`
	Email    string             `json:"email" binding:"required"`
}

type Feature struct {
	F string `json:"feature"`
}

type Product struct {
	ID          primitive.ObjectID `json:"_id"`
	Name        string             `json:"name"`
	Price       float64            `json:"price"`
	Currency    string             `json:"currency"`
	Quantity    int64              `json:"quantity"`
	Description string             `json:"description"`
	Category    string             `json:"category"`
	Features    []Feature          `json:"features"`
	Reviews     []Review           `json:"reviews"`
	CreatedAt   time.Time          `json:"created_at"`
}

type Review struct {
	User    string `json:"user"`
	Stars   int8   `json:"stars"`
	Comment string `json:"comment"`
}
