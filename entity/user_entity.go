package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateUserRequest models the create user request json body
type CreateUserRequest struct {
	Username     string    `json:"username" binding:"required"`
	Fullname     string    `json:"fullname" binding:"required"`
	Email        string    `json:"email" binding:"email,required"`
	Password     string    `json:"password" binding:"required"`
	MobileNumber string    `json:"mobile_number"`
	CreatedAt    time.Time `json:"created_at"`
}

type User struct {
	ID                      primitive.ObjectID `json:"_id" bson:"_id"`
	Username                string             `json:"username" bson:"username" binding:"required"`
	PasswordSalt            string             `json:"salt" bson:"salt"`
	HashedPassword          string             `json:"password" bson:"password" binding:"required"`
	Fullname                string             `json:"fullname" bson:"fullname" binding:"required"`
	Email                   string             `json:"email" bson:"email" binding:"email,required"`
	MobileNumber            string             `json:"mobile_number" bson:"mobile_number"`
	ProfilePicture          string             `json:"picture" bson:"picture"`
	CreatedAt               time.Time          `json:"created_at" bson:"created_at"`
	LastUpdated             time.Time          `json:"last_updated" bson:"last_updated"`
	EmailIsVerfied          bool               `json:"email_is_verified" bson:"email_is_verified"`
	DefaultPaymentMethod    string             `json:"default_payment_method" bson:"default_payment_method"`
	SavedPaymentDetails     string             `json:"saved_payment_details" bson:"saved_payment_details"`
	Orders                  []Order            `json:"orders" bson:"orders"`
	DefaultDeliveryLocation Location           `json:"default_delivery_location" bson:"default_delivery_loaction"`

	// Optional
	FavouriteProducts   []Product  `json:"favourite_products,omitempty" bson:"favourite_products"`
	RegisteredLocations []Location `json:"locations,omitempty" bson:"locations"`
}

type Location struct {
	HouseNumber string `json:"house_number,omitempty"`
	PhoneNo     string `json:"telephone,omitempty"`
	Street      string `json:"street,omitempty"`
	CityOrTown  string `json:"city_or_town,omitempty"`
	State       string `json:"state,omitempty"`
	Country     string `json:"country,omitempty"`
	ZipCode     string `json:"zip_code,omitempty"`
}
