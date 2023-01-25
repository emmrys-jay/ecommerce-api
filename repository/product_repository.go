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

func InsertOneProduct(collection *mongo.Collection, data interface{}) (*mongo.InsertOneResult, error) {
	return collection.InsertOne(context.Background(), data)
}

func InsertProducts(collection *mongo.Collection, data []entity.Product) (*mongo.InsertManyResult, error) {
	var ui []interface{}
	for _, t := range data {
		ui = append(ui, t)
	}

	return collection.InsertMany(context.Background(), ui)
}

func FindOneProduct(collection *mongo.Collection, productID string) (*entity.Product, error) {
	ctx := context.Background()
	var product = entity.Product{}

	filter := bson.M{"_id": productID}
	err := collection.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func FindProducts(collection *mongo.Collection, name string, offset, limit int64) ([]entity.Product, int64, error) {
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

	length, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, -1, err
	}

	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(offset)

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, -1, err
	}
	defer cursor.Close(context.Background())

	var products = []entity.Product{}
	for cursor.Next(ctx) {
		var product entity.Product
		_ = cursor.Decode(&product)
		products = append(products, product)
	}

	return products, length, nil
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

// UpdateProduct updates a product price/ quantity and orders
func UpdateProduct(collection *mongo.Collection, id string, price float64, quantity int64, productOrders int64) (*mongo.UpdateResult, error) {
	ctx := context.Background()

	product, err := FindOneProduct(collection, id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": id}

	if price != 0 {
		product.Price = price
		product.LastUpdated = time.Now()
	}

	if quantity != 0 {
		product.Quantity += quantity
		product.LastUpdated = time.Now()
	}

	if productOrders != 0 {
		product.NumOfOrders += productOrders
		product.LastUpdated = time.Now()
	}

	result, err := collection.ReplaceOne(ctx, filter, product)

	return result, err
}

func AddProductReview(collection *mongo.Collection, productID string, review entity.Review) (*mongo.UpdateResult, error) {
	ctx := context.Background()

	product, err := FindOneProduct(collection, productID)
	if err != nil {
		return nil, err
	}

	review.CreatedAt = time.Now()
	product.LastUpdated = time.Now()
	product.Reviews = append(product.Reviews, review)

	filter := bson.M{"_id": productID}

	result, err := collection.ReplaceOne(ctx, filter, product)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetProductsByCategory(collection *mongo.Collection, ctgy string, offset, limit int) ([]entity.Product, int64, error) {
	ctx := context.Background()
	var products = []entity.Product{}
	var product entity.Product

	filter := bson.M{"category": ctgy}

	length, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, -1, err
	}

	options := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, filter, options)
	if err != nil {
		return nil, -1, err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(ctx) {
		err := cursor.Decode(&product)
		if err != nil {
			return nil, -1, err
		}

		products = append(products, product)
	}

	return products, length, nil
}
