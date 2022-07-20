package repository

import (
	"context"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func OrderProductDirectly(collection *mongo.Collection, location entity.Location, quantity int, username, fullname, productName, paymentMethod string) (*mongo.InsertOneResult, error) {
	ctx := context.Background()

	product, err := FindOneProduct(collection, productName)
	if err != nil {
		return nil, err
	}

	order := entity.Order{
		ID:               primitive.NewObjectIDFromTimestamp(time.Now()),
		Username:         username,
		FullName:         fullname,
		DeliveryLocation: location,
		Product:          product,
		ProductQuantity:  quantity,
		IsDelivered:      false,
		CreatedAt:        time.Now(),
	}

	result, err := collection.InsertOne(ctx, order)
	if result != nil {
		return nil, err
	}

	return result, nil
}

func GetSingleOrder(collection *mongo.Collection, username string, id primitive.ObjectID) (*entity.Order, error) {
	ctx := context.Background()
	var order entity.Order

	filter := bson.M{
		"$and": []bson.M{
			{
				"username": username,
			},
			{
				"_id": id,
			},
		},
	}

	result := collection.FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err()
	}

	err := result.Decode(&order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func GetOrdersWithUsername(collection *mongo.Collection, username string) ([]entity.Order, error) {
	ctx := context.Background()
	var orders = []entity.Order{}
	var order = entity.Order{}

	filter := bson.M{"username": username}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	for cursor.Next(ctx) {
		err := cursor.Decode(&order)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func GetAllOrders(collection *mongo.Collection, limit, offset int) ([]entity.Order, error) {
	ctx := context.Background()
	var order = entity.Order{}
	var orders = []entity.Order{}

	filter := bson.M{}

	options := options.Find()
	options.SetLimit(int64(limit))
	options.SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}

	for cursor.Next(ctx) {
		err := cursor.Decode(&order)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}
