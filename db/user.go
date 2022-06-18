package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateUser(collection *mongo.Collection, user CreateUserRequest) (*mongo.InsertOneResult, error) {
	ctx := context.Background()

	user.CreatedAt = time.Now()

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetUser(collection *mongo.Collection, username string) (*User, error) {
	ctx := context.Background()
	var user = &User{}
	filter := bson.M{"username": username}

	result := collection.FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, mongo.ErrNoDocuments
	}

	err := result.Decode(user)

	return user, err
}

func GetAllUsers(collection *mongo.Collection) ([]User, error) {
	ctx := context.Background()
	var user = User{}
	var users []User
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
