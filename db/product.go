package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InsertOneProduct(collection *mongo.Collection, data interface{}) (*mongo.InsertOneResult, error) {
	return collection.InsertOne(context.Background(), data)
}

func InsertProducts(collection *mongo.Collection, data []Product) (*mongo.InsertManyResult, error) {
	var ui []interface{}
	for _, t := range data {
		ui = append(ui, t)
	}

	return collection.InsertMany(context.Background(), ui)
}

func FindOneProduct(collection *mongo.Collection, name string) (Product, error) {
	ctx := context.Background()
	var product Product

	filter := bson.M{"name": name}
	err := collection.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		return product, err
	}

	return product, nil
}

func FindProducts(collection *mongo.Collection, name string, offset, limit int64) ([]Product, error) {
	ctx := context.Background()
	filter := bson.M{}
	findOptions := options.Find()
	//name = "iphone"
	if name != "" {
		filter = bson.M{
			"$or": []bson.M{
				{
					"name": bson.M{
						"$regex": primitive.Regex{
							Pattern: name,
							Options: "i",
						},
					},
				},
				{
					"description": bson.M{
						"$regex": primitive.Regex{
							Pattern: name,
							Options: "i",
						},
					},
				},
			},
		}
	}

	findOptions.SetLimit(int64(limit))

	if offset != 0 {
		findOptions.SetSkip(2)
	}

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var products = []Product{}
	for cursor.Next(ctx) {
		var product Product
		_ = cursor.Decode(&product)
		products = append(products, product)
	}

	return products, nil
}

func DeleteProduct(collection *mongo.Collection, name string) (*mongo.DeleteResult, error) {
	ctx := context.Background()
	filter := bson.D{{Key: "name", Value: name}}

	result, err := collection.DeleteOne(ctx, filter)
	return result, err
}

func DeleteAllProducts(collection *mongo.Collection) (*mongo.DeleteResult, error) {
	ctx := context.Background()
	filter := bson.M{}

	result, err := collection.DeleteMany(ctx, filter)
	return result, err
}

// UpdateProduct updates a product price and/or quantity
func UpdateProduct(collection *mongo.Collection, name string, price float64, quantity int64) (*mongo.UpdateResult, error) {
	ctx := context.Background()

	product, err := FindOneProduct(collection, name)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"name": name}

	if price != 0 {
		product.Price = price
		product.LastUpdated = time.Now()
	}

	if quantity != 0 {
		product.Quantity += quantity
		product.LastUpdated = time.Now()
	}

	result, err := collection.ReplaceOne(ctx, filter, product)

	return result, err
}

func AddProductReview(collection *mongo.Collection, name string, review Review) (*mongo.UpdateResult, error) {
	ctx := context.Background()

	product, err := FindOneProduct(collection, name)
	if err != nil {
		return nil, err
	}

	review.CreatedAt = time.Now()
	product.LastUpdated = time.Now()
	product.Reviews = append(product.Reviews, review)

	filter := bson.M{"name": name}

	result, err := collection.ReplaceOne(ctx, filter, product)
	if err != nil {
		return nil, err
	}

	return result, nil
}
