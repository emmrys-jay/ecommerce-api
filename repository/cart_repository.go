package repository

import (
	"context"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddToCart(collection *mongo.Collection, quantity int64, productName, username string) (*mongo.InsertOneResult, error) {
	ctx := context.Background()

	product, err := FindOneProduct(db.GetCollection(collection.Database(), "products"), productName)
	if err != nil {
		return nil, err
	}

	item := entity.CartItem{
		ID:        primitive.NewObjectIDFromTimestamp(time.Now()),
		Username:  username,
		Quantity:  quantity,
		DateAdded: time.Now(),
		Product:   product,
	}

	result, err := collection.InsertOne(ctx, item)

	return result, err
}

func RemoveFromCart(collection *mongo.Collection, productName, username string) (*mongo.DeleteResult, error) {
	ctx := context.Background()

	filter := bson.M{
		"$and": []bson.M{
			{
				"product_name": productName,
			},
			{
				"username": username,
			},
		},
	}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func SubtractCartQuantity(collection *mongo.Collection, productName, username string) (*mongo.UpdateResult, error) {
	ctx := context.Background()

	filter := bson.M{
		"$and": []bson.M{
			{
				"product_name": productName,
			},
			{
				"username": username,
			},
		},
	}

	cartItem, err := GetCartItem(collection, productName, username)
	if err != nil {
		return nil, err
	}

	cartItem.Quantity -= 1

	result, err := collection.ReplaceOne(ctx, filter, cartItem)

	return result, err
}

func GetCartItem(collection *mongo.Collection, productName, username string) (*entity.CartItem, error) {
	ctx := context.Background()
	var cartItem = entity.CartItem{}

	filter := bson.M{
		"$and": []bson.M{
			{
				"product_name": productName,
			},
			{
				"username": username,
			},
		},
	}

	result := collection.FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err()
	}

	err := result.Decode(&cartItem)
	return &cartItem, err
}

func GetUserCartItems(collection *mongo.Collection, username string) ([]entity.CartItem, error) {
	ctx := context.Background()
	var cartItems = []entity.CartItem{}

	filter := bson.M{
		"username": username,
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	for cursor.Next(ctx) {
		var cartItem entity.CartItem
		err := cursor.Decode(&cartItem)
		if err != nil {
			return nil, err
		}
		cartItems = append(cartItems, cartItem)
	}

	return cartItems, err
}

func DeleteAllCartItems(collection *mongo.Collection) (*mongo.DeleteResult, error) {
	ctx := context.Background()

	filter := bson.M{}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return nil, err
	}

	return result, err
}
