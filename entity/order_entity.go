package entity

import (
	"time"
)

type Order struct {
	ID               string    `json:"_id" bson:"_id"`
	UserID           string    `json:"user_id,omitempty" bson:"user_id" binding:"required" description:"user who placed the order"`
	FullName         string    `json:"fullname,omitempty" bson:"fullname" binding:"required" description:"fullname specified during checkout"`
	DeliveryLocation Location  `json:"delivery_address,omitempty" bson:"delivery_address" description:"location specified during checkout"`
	DeliveryFee      float64   `json:"delivery_fee,omitempty" bson:"delivery_fee" binding:"required"`
	Product          Product   `json:"product,omitempty" bson:"product" binding:"required"`
	ProductQuantity  int       `json:"product_quantity,omitempty" bson:"product_quantity" binding:"required"`
	IsDelivered      bool      `json:"is_delivered,omitempty" bson:"is_delivered"  binding:"required"`
	CreatedAt        time.Time `json:"created_at,omitempty" bson:"created_at"`
	TimeDelivered    time.Time `json:"time_delivered,omitempty" bson:"time_delivered"`
	IsReceived       bool      `json:"is_received,omitempty" bson:"is_received"`
}
