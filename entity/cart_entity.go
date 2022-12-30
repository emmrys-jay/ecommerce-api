package entity

import (
	"time"
)

type CartItem struct {
	ID        string    `json:"_id" bson:"_id"`
	ProductID string    `json:"product_id" bson:"product_id"`
	UserID    string    `json:"user_id" bson:"user_id"`
	Quantity  int64     `json:"quantity" bson:"quantity"`
	DateAdded time.Time `json:"date_added" bson:"date_added"`
	Product   Product   `json:"product" bson:"product"`
}
