package entity

import (
	"time"
)

type Product struct {
	ID          string    `json:"_id" bson:"_id"`
	Name        string    `json:"name,omitempty" bson:"name" binding:"required"`
	Price       float64   `json:"price,omitempty" bson:"price" binding:"required"`
	Pictures    []string  `json:"pictures" bson:"pictures"`
	Videos      []string  `json:"videos" bson:"videos"`
	Currency    string    `json:"currency,omitempty" bson:"currency"`
	Quantity    int64     `json:"quantity,omitempty" bson:"quantity" binding:"required"`
	Description string    `json:"description,omitempty" bson:"description" binding:"required"`
	Category    string    `json:"category,omitempty" bson:"category" binding:"required"`
	Features    []Feature `json:"features,omitempty" bson:"features"`
	Reviews     []Review  `json:"reviews,omitempty" bson:"reviews"`
	NoOfReviews int64     `json:"no_of_reviews,omitempty" bson:"no_of_reviews"`
	CreatedAt   time.Time `json:"created_at,omitempty" bson:"created_at"`
	LastUpdated time.Time `json:"last_updated,omitempty" bson:"last_updated"`
	NumOfOrders int64     `json:"num_of_orders,omitempty"`

	// Optional
	SlashedPrice float64 `json:"slashed_price,omitempty" bson:"slashed_price"`
	MinimumOrder int64   `json:"minimum_order,omitempty"`
}

type Feature struct {
	F string `json:"feature" bson:"feature"`
}

type Review struct {
	User      string    `json:"user"`
	Stars     int64     `json:"stars" binding:"required,min=1,max=5"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
