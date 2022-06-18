package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateUserRequest models the create user request json body
type CreateUserRequest struct {
	Username  string    `json:"username" binding:"required"`
	Fullname  string    `json:"fullname" binding:"required"`
	Email     string    `json:"email" binding:"email,required"`
	Password  string    `json:"password" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id"`
	Username       string             `json:"username" bson:"username" binding:"required"`
	HashedPassword string             `json:"password" bson:"password" binding:"required"`
	Fullname       string             `json:"fullname" bson:"fullname" binding:"required"`
	Email          string             `json:"email" bson:"email" binding:"email,required"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
}

type Feature struct {
	F string `json:"feature"`
}

type Product struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	Name        string             `json:"name" bson:"name" binding:"required"`
	Price       float64            `json:"price" bson:"price" binding:"required"`
	Currency    string             `json:"currency" bson:"currency" binding:"required"`
	Quantity    int64              `json:"quantity" bson:"quantity" binding:"required"`
	Description string             `json:"description" bson:"description" binding:"required"`
	Category    string             `json:"category" bson:"category" binding:"required"`
	Features    []Feature          `json:"features" bson:"features"`
	Reviews     []Review           `json:"reviews" bson:"reviews"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	LastUpdated time.Time          `json:"last_updated" bson:"last_updated"`
}

type Review struct {
	User      string    `json:"user" binding:"required"`
	Stars     int8      `json:"stars" binding:"required"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}
