package entity

import (
	"time"
)

// PaginationResponse models the response of a GET request that enables pagination
type PaginationResponse struct {
	PageID        int `json:"page_id"`
	ResultsFound  int `json:"results_found"`
	NumberOfPages int `json:"no_of_pages"`
	Data          any `json:"data"`
}

// CreateUserRequest models the create user request json body
type CreateUserRequest struct {
	Username     string `json:"username" binding:"required"`
	Fullname     string `json:"fullname" binding:"required"`
	Email        string `json:"email" binding:"email,required"`
	Password     string `json:"password" binding:"required"`
	MobileNumber string `json:"mobile_number"`
}

type User struct {
	ID                      string    `json:"_id" bson:"_id"`
	Username                string    `json:"username" bson:"username" binding:"required"`
	PasswordSalt            string    `json:"-" bson:"salt"`
	Password                string    `json:"-" bson:"password" binding:"required"`
	Fullname                string    `json:"fullname" bson:"fullname" binding:"required"`
	Email                   string    `json:"email" bson:"email" binding:"email,required"`
	MobileNumber            string    `json:"mobile_number" bson:"mobile_number"`
	ProfilePicture          string    `json:"picture" bson:"picture"`
	CreatedAt               time.Time `json:"created_at" bson:"created_at"`
	LastUpdated             time.Time `json:"last_updated" bson:"last_updated"`
	EmailIsVerfied          bool      `json:"email_is_verified" bson:"email_is_verified"`
	DefaultPaymentMethod    string    `json:"default_payment_method" bson:"default_payment_method"`
	SavedPaymentDetails     string    `json:"saved_payment_details" bson:"saved_payment_details"`
	DefaultDeliveryLocation Location  `json:"default_delivery_location" bson:"default_delivery_loaction"`

	// Optional
	FavouriteProducts   []string   `json:"favourite_products,omitempty" bson:"favourite_products" description:"ID's of user's favourite products"`
	RegisteredLocations []Location `json:"locations,omitempty" bson:"locations"`
}

// UserResponse models the response of a createuser or loginuser request
type UserResponse struct {
	ID             string    `json:"_id"`
	Username       string    `json:"username"`
	Fullname       string    `json:"fullname"`
	Email          string    `json:"email"`
	Token          string    `json:"token"`
	CreatedAt      time.Time `json:"created_at"`
	EmailIsVerfied bool      `json:"email_is_verified"`
	MobileNumber   string    `json:"mobile_number,omitempty"`
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
