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

type Feature struct {
	F string `json:"feature" bson:"feature"`
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

type Product struct {
	// TODO FIx this issue with the _id json not showing in the json result

	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	Name        string             `json:"name,omitempty" bson:"name" binding:"required"`
	Price       float64            `json:"price,omitempty" bson:"price" binding:"required"`
	Pictures    []string           `json:"pictures" bson:"pictures"`
	Videos      []string           `json:"videos" bson:"videos"`
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
	SlashedPrice float64 `json:"slashed_price,omitempty" bson:"slashed_price"`
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
	DeliveryFee      float64            `json:"delivery_fee,omitempty" binding:"required"`
	Product          Product            `json:"product,omitempty" binding:"required"`
	ProductQuantity  int                `json:"product_quantity,omitempty" binding:"required"`
	IsDelivered      bool               `json:"is_delivered,omitempty" binding:"required"`
	CreatedAt        time.Time          `json:"created_at,omitempty"`
	TimeDelivered    time.Time          `json:"time_delivered,omitempty"`

	// Optional
	OrderID    string `json:"order_id" bson:"order_id"`
	IsReceived bool   `json:"is_received,omitempty"`
}
