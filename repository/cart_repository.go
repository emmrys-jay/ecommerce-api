package repository

import (
	"context"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddToCart(collection *mongo.Collection, quantity int64, productID, username string) (*mongo.InsertOneResult, error) {
	ctx := context.Background()

	product, err := FindOneProduct(db.GetCollection(collection.Database(), "products"), productID)
	if err != nil {
		return nil, err
	}

	item := entity.CartItem{
		ID:        primitive.NewObjectIDFromTimestamp(time.Now()).String()[10:34],
		Username:  username,
		Quantity:  quantity,
		DateAdded: time.Now(),
		Product:   *product,
	}

	result, err := collection.InsertOne(ctx, item)

	return result, err
}

func RemoveFromCart(collection *mongo.Collection, cartItemID, username string) (*mongo.DeleteResult, error) {
	ctx := context.Background()

	filter := bson.M{
		"$and": []bson.M{
			{
				"_id": cartItemID,
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

func UpdateCartQuantity(collection *mongo.Collection, quantity int, cartItemID, username string) (*mongo.UpdateResult, error) {
	ctx := context.Background()

	filter := bson.M{
		"$and": []bson.M{
			{
				"_id": cartItemID,
			},
			{
				"username": username,
			},
		},
	}

	cartItem, err := GetCartItem(collection, cartItemID, username)
	if err != nil {
		return nil, err
	}

	cartItem.Quantity = int64(quantity)

	result, err := collection.ReplaceOne(ctx, filter, cartItem)

	return result, err
}

func GetCartItem(collection *mongo.Collection, cartItemID, username string) (*entity.CartItem, error) {
	ctx := context.Background()
	var cartItem = entity.CartItem{}

	filter := bson.M{"_id": cartItemID}

	if username != "" {
		filter = bson.M{
			"$and": []bson.M{
				{
					"_id": cartItemID,
				},
				{
					"username": username,
				},
			},
		}
	}

	result := collection.FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err()
	}

	err := result.Decode(&cartItem)
	return &cartItem, err
}

func GetUserCartItems(collection *mongo.Collection, username string, offset, limit int) ([]entity.CartItem, int64, error) {
	ctx := context.Background()
	var cartItems = []entity.CartItem{}
	option := options.Find()
	var length int64
	var err error

	filter := bson.M{
		"username": username,
	}

	if offset != 0 || limit != 0 {
		length, err = collection.CountDocuments(ctx, filter)
		if err != nil {
			return nil, 0, err
		}

		option = option.SetLimit(int64(limit)).SetSkip(int64(offset))
	}

	cursor, err := collection.Find(ctx, filter, option)
	if err != nil {
		return nil, 0, err
	}

	for cursor.Next(ctx) {
		var cartItem entity.CartItem
		err := cursor.Decode(&cartItem)
		if err != nil {
			return nil, 0, err
		}
		cartItems = append(cartItems, cartItem)
	}

	return cartItems, length, err
}

func GetAllCartItems(collection *mongo.Collection, offset, limit int) ([]entity.CartItem, int64, error) {
	ctx := context.Background()
	var cartItems = []entity.CartItem{}
	option := options.Find()
	var length int64
	var err error

	filter := bson.M{}

	length, err = collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	option = option.SetLimit(int64(limit)).SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, filter, option)
	if err != nil {
		return nil, 0, err
	}

	for cursor.Next(ctx) {
		var cartItem entity.CartItem
		err := cursor.Decode(&cartItem)
		if err != nil {
			return nil, 0, err
		}
		cartItems = append(cartItems, cartItem)
	}

	return cartItems, length, err
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

func DeleteAllUserCartItems(collection *mongo.Collection, username string) (*mongo.DeleteResult, error) {
	ctx := context.Background()

	filter := bson.M{"username": username}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}

	return result, err
}

func DeleteCartItem(collection *mongo.Collection, id string) (*mongo.DeleteResult, error) {
	ctx := context.Background()

	filter := bson.M{"_id": id}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}

	return result, err
}
