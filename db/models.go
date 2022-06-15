package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID `json:"_id"`
	Username       string             `json:"username" binding:"required"`
	HashedPassword string             `json:"password" binding:"required"`
	Fullname       string             `json:"fullname" binding:"required"`
	Email          string             `json:"email" binding:"email,required"`
	Token          string             `json:"token" binding:"required"`
	CreatedAt      time.Time          `json:"created_at" binding:"required"`
}

type Feature struct {
	F string `json:"feature"`
}

type Product struct {
	ID          primitive.ObjectID `json:"_id"`
	Name        string             `json:"name" binding:"required"`
	Price       float64            `json:"price" binding:"required"`
	Currency    string             `json:"currency" binding:"required"`
	Quantity    int64              `json:"quantity" binding:"required"`
	Description string             `json:"description" binding:"required"`
	Category    string             `json:"category" binding:"required"`
	Features    []Feature          `json:"features"`
	Reviews     []Review           `json:"reviews"`
	CreatedAt   time.Time          `json:"created_at"`
	LastUpdated time.Time          `json:"last_updated"`
}

type Review struct {
	User      string    `json:"user" binding:"required"`
	Stars     int8      `json:"stars" binding:"required"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}
