package db

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
	Username                string             `json:"username,omitempty" bson:"username" binding:"required"`
	PasswordSalt            string             `json:"salt,omitempty"`
	HashedPassword          string             `json:"password,omitempty" bson:"password" binding:"required"`
	Fullname                string             `json:"fullname,omitempty" bson:"fullname" binding:"required"`
	Email                   string             `json:"email,omitempty" bson:"email" binding:"email,required"`
	MobileNumber            string             `json:"mobile_number,omitempty"`
	ProfilePicture          string             `json:"picture,omitempty"`
	CreatedAt               time.Time          `json:"created_at,omitempty" bson:"created_at"`
	LastUpdated             time.Time          `json:"last_updated,omitempty" bson:"last_updated"`
	EmailIsVerfied          bool               `json:"email_is_verified,omitempty" bson:"email_is_verified"`
	DefaultPaymentMethod    string             `json:"default_payment_method,omitempty"`
	SavedPaymentDetails     string             `json:"saved_payment_details,omitempty"`
	Orders                  []Order            `json:"orders,omitempty"`
	DefaultDeliveryLocation Location           `json:"default_delivery_location,omitempty"`

	// Optional
	FavouriteProducts   []Product  `json:"favourite_products,omitempty"`
	RegisteredLocations []Location `json:"locations,omitempty"`
}

type Feature struct {
	F string `json:"feature"`
}

type Location struct {
	HouseNumber string `json:"house_number,omitempty"`
	Street      string `json:"street,omitempty" binding:"required"`
	CityOrTown  string `json:"city_or_town,omitempty" binding:"required"`
	State       string `json:"state,omitempty" binding:"required"`
	Country     string `json:"country,omitempty" binding:"required"`
	ZipCode     string `json:"zip_code,omitempty"`
}

type Product struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	Name        string             `json:"name,omitempty" bson:"name" binding:"required"`
	Price       float64            `json:"price,omitempty" bson:"price" binding:"required"`
	Pictures    []string           `json:"pictures" bson:"pictures"`
	Videos      []string           `json:"videos" bson:"prictures"`
	Currency    string             `json:"currency,omitempty" bson:"currency"`
	Quantity    int64              `json:"quantity,omitempty" bson:"quantity" binding:"required"`
	Description string             `json:"description,omitempty" bson:"description" binding:"required"`
	Category    string             `json:"category,omitempty" bson:"category" binding:"required"`
	Features    []Feature          `json:"features,omitempty" bson:"features"`
	Reviews     []Review           `json:"reviews,omitempty" bson:"reviews"`
	CreatedAt   time.Time          `json:"created_at,omitempty" bson:"created_at"`
	LastUpdated time.Time          `json:"last_updated,omitempty" bson:"last_updated"`
	NumOfOrders int64              `json:"num_of_orders,omitempty"`

	// Optional
	SlashedPrice float64 `json:"slashed_price,omitempty" bson:"price"`
	MinimumOrder int64   `json:"minimum_order,omitempty"`
}

type Review struct {
	User      string    `json:"user" binding:"required"`
	Stars     int64     `json:"stars" binding:"required"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CartItem struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	ProductName string             `json:"product_name" bson:"product_name"`
	Username    string             `json:"username" bson:"username"`
	Quantity    int64              `json:"quantity" bson:"quantity"`
	DateAdded   time.Time          `json:"date_added" bson:"date_added"`
	Product     Product            `json:"product" bson:"product"`
}

type Order struct {
	ID               primitive.ObjectID `json:"_id" bson:"_id"`
	Username         string             `json:"username,omitempty" binding:"required" description:"user who placed the order"`
	FullName         string             `json:"fullname,omitempty" binding:"required" description:"fullname specified during checkout"`
	DeliveryLocation Location           `json:"delivery_address,omitempty" description:"location specified during checkout"`
	DeliveryPhone    string             `json:"delivery_phone,omitempty" binding:"required"`
	DeliveryFee      float64            `json:"delivery_fee,omitempty" binding:"required"`
	Product          Product            `json:"product,omitempty" binding:"required"`
	IsDelivered      bool               `json:"is_delivered,omitempty" binding:"required"`
	CreatedAt        time.Time          `json:"created_at,omitempty"`
	TimeDelivered    time.Time          `json:"time_delivered,omitempty"`

	// Optional
	OrderID    string `json:"order_id" bson:"order_id"`
	IsReceived bool   `json:"is_received,omitempty"`
}
