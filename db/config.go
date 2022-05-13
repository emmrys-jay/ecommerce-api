package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//const dbDetails = "mongodb://ecommerce-api:ecommerceapp001@localhost:27017"
var DB *mongo.Database

func ConfigDB() {
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

	DB = client.Database("ecommerce")
}
