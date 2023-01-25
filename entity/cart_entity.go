package entity

import (
	"time"
)

type CartItem struct {
	ID          string    `json:"_id" bson:"_id"`
	ProductName string    `json:"product_name" bson:"product_name"`
	Username    string    `json:"username" bson:"username"`
	Quantity    int64     `json:"quantity" bson:"quantity"`
	DateAdded   time.Time `json:"date_added" bson:"date_added"`
	Product     Product   `json:"product" bson:"product"`
}
