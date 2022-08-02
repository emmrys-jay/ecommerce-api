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
		log.Fatalln("could not connect to server: ", err)
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalln("could not ping server: ", err)
	}

	fmt.Println("You connected to your mongo database.")

	db := client.Database("ecommerce")
	ConfigDBCollections(db)

	return db
}

func GetCollection(db *mongo.Database, collection string) *mongo.Collection {
	return db.Collection(collection)
}

func ConfigDBCollections(db *mongo.Database) {
	collection := GetCollection(db, "users")
	ctx := context.Background()

	options := options.Index()
	options.SetUnique(true)

	collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: "1"}, {Key: "email", Value: "1"}, {Key: "mobile_number", Value: "1"}},
		Options: options,
	})

	collection = GetCollection(db, "products")

	collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: "1"}},
		Options: options,
	})

	// collection = GetCollection(db, "cart")

	// collection.Indexes().CreateOne(ctx, mongo.IndexModel{
	// 	Keys:    bson.D{{Key: "product_name", Value: "1"}},
	// 	Options: options,
	// })
}
