package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartItem struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	ProductName string             `json:"product_name" bson:"product_name"`
	Username    string             `json:"username" bson:"username"`
	Quantity    int64              `json:"quantity" bson:"quantity"`
	DateAdded   time.Time          `json:"date_added" bson:"date_added"`
	Product     Product            `json:"product" bson:"product"`
}
