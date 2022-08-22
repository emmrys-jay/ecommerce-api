package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//const dbDetails = "mongodb://ecommerce-api:ecommerceapp001@localhost:27017"

// Add username param and password
func ConfigDB() *mongo.Database {
	// get a mongo sessions
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalln("mongo.Connect(ctx, options.Client().ApplyURI() ERROR: ", err)
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalln("client.Ping(ctx, readpref.Primary()) ERROR: ", err)
	}

	fmt.Println("You connected to your mongo database.")

	db := client.Database("ecommerce")
	err = ConfigDBCollections(db)
	if err != nil {
		log.Fatalln("ConfigDBCOllections() ERROR: ", err)
	}

	return db
}

func GetCollection(db *mongo.Database, collection string) *mongo.Collection {
	return db.Collection(collection)
}

func ConfigDBCollections(db *mongo.Database) error {
	collection := GetCollection(db, "users")
	ctx := context.Background()

	_, err := collection.Indexes().CreateMany(ctx,
		[]mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "username", Value: 1}},
				Options: options.Index().SetName("username_index").SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetName("email_index").SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "mobile_number", Value: 1}},
				Options: options.Index().SetName("mobile_number_index").SetUnique(true),
			},
		})

	if err != nil {
		return err
	}

	collection = GetCollection(db, "products")

	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetName("name_index").SetUnique(true),
	})

	if err != nil {
		return err
	}

	collection = GetCollection(db, "cart")

	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "product_name", Value: 1}},
		Options: options.Index().SetName("product_name_index").SetUnique(true),
	})

	return err
}
