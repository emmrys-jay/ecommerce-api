package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateUser(collection *mongo.Collection, user entity.User) (*mongo.InsertOneResult, error) {
	ctx := context.Background()

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetUser(collection *mongo.Collection, username string) (*entity.User, error) {
	ctx := context.Background()
	var user = &entity.User{}
	filter := bson.M{"username": username}

	result := collection.FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, mongo.ErrNoDocuments
	}

	err := result.Decode(user)

	return user, err
}

func GetAllUsers(collection *mongo.Collection) ([]entity.User, error) {
	ctx := context.Background()
	var user = entity.User{}
	var users []entity.User
	filter := bson.M{}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	for cursor.Next(ctx) {
		cursor.Decode(&user)
		users = append(users, user)
	}

	return users, nil
}

func DeleteUser(collection *mongo.Collection, username string) (*mongo.DeleteResult, error) {
	ctx := context.Background()

	filter := bson.M{"username": username}

	result, err := collection.DeleteOne(ctx, filter)

	return result, err
}

func DeleteAllUsers(collection *mongo.Collection) (*mongo.DeleteResult, error) {
	ctx := context.Background()

	filter := bson.M{}

	result, err := collection.DeleteMany(ctx, filter)

	return result, err
}

/*
* Models direct database querying functions to update/modify the following user fields:
* - username
* - password
* - profile_picture
* - email
* - mobile number
* - default payment method
 */
func UpdateUserFlexible(collection *mongo.Collection, username, detail, update, salt string) error {
	ctx := context.Background()
	var user = entity.User{}

	filter := bson.M{"username": username}
	result := collection.FindOne(ctx, filter)

	err := result.Decode(&user)
	if err != nil {
		return err
	}

	switch detail {
	case "username":
		user.Username = update
	case "password":
		user.HashedPassword = update
		user.PasswordSalt = salt
	case "profile_picture":
		user.ProfilePicture = update
	case "email":
		user.Email = update
	case "mobile_number":
		user.MobileNumber = update
	case "default_payment_method":
		user.DefaultPaymentMethod = update
	default:
		return fmt.Errorf("invalid field specified")
	}

	user.LastUpdated = time.Now()

	_, err = collection.ReplaceOne(ctx, filter, user)

	return err
}

func AddLocation(collection *mongo.Collection, username string, location entity.Location) (*entity.Location, error) {
	ctx := context.Background()

	filter := bson.M{"username": username}
	user, err := GetUser(collection, username)
	if err != nil {
		return nil, err
	}

	if user.DefaultDeliveryLocation.CityOrTown == "" {
		user.DefaultDeliveryLocation = location
	}

	user.RegisteredLocations = append(user.RegisteredLocations, location)
	user.LastUpdated = time.Now()

	_, err = collection.ReplaceOne(ctx, filter, user)

	return &location, err
}

func AddOrderToUser(collection *mongo.Collection, username string, order *entity.Order) (*mongo.UpdateResult, error) {
	ctx := context.Background()

	filter := bson.M{"username": username}

	user, err := GetUser(collection, username)
	if err != nil {
		return nil, err
	}

	user.Orders = append(user.Orders, *order)
	user.LastUpdated = time.Now()

	result, err := collection.ReplaceOne(ctx, filter, user)

	return result, err

}

// VerifyEmail function
