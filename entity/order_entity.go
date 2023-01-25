package entity

import (
	"time"
)

type Order struct {
	ID               string    `json:"_id" bson:"_id"`
	Username         string    `json:"username,omitempty" binding:"required" description:"user who placed the order"`
	FullName         string    `json:"fullname,omitempty" binding:"required" description:"fullname specified during checkout"`
	DeliveryLocation Location  `json:"delivery_address,omitempty" description:"location specified during checkout"`
	DeliveryFee      float64   `json:"delivery_fee,omitempty" binding:"required"`
	Product          Product   `json:"product,omitempty" binding:"required"`
	ProductQuantity  int       `json:"product_quantity,omitempty" binding:"required"`
	IsDelivered      bool      `json:"is_delivered,omitempty" binding:"required"`
	CreatedAt        time.Time `json:"created_at,omitempty"`
	TimeDelivered    time.Time `json:"time_delivered,omitempty"`

	// Optional
	IsReceived bool `json:"is_received,omitempty"`
}
