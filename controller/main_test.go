package controller

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type ServerDB struct {
	Db     *mongo.Database
	Server *gin.Engine
}

func NewServerDB() *ServerDB {
	return &ServerDB{
		Db:     connectDB(),
		Server: gin.New(),
	}
}

func connectDB() *mongo.Database {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalln("could not connect to server: ", err)
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalln("could not ping server: ", err)
	}

	database := client.Database("ecommerce_test")
	fmt.Println("Successfully connected to database")
	return database
}

func initializeUserRoutes(ed *ServerDB) {
	userController := NewUserController(ed.Db)

	user := ed.Server.Group("/user")
	{
		user.POST("/create", userController.CreateUser)
		user.POST("/login", userController.LoginUser)
		// user.POST("/logout", userController.LogoutUser)
		user.GET("/get", userController.GetUser)
		user.PUT("/password", userController.ChangePassword)
		user.PUT("/update", userController.UpdateUserFlexible)
		user.PUT("/location/add", userController.AddLocation)
	}
}

func initializeProductRoutes(details *ServerDB) {
	userController := NewUserController(details.Db)

	products := details.Server.Group("/products")
	{
		products.GET("/get/:category", userController.GetProductsByCategory)
		products.GET("/find", userController.FindProducts)
		products.GET("/findone/:productID", userController.FindOneProduct)
		// products.GET("/find/recent", userController.FindProductsWithTime)
		//products.GET("/find/reviews", userController.FindProductsBasedOnReviews)
		products.PUT("/:productID/addreview", userController.AddReview)
		// products.GET("/categories", getAllCategories)
	}
}

func initializeOrdersRoutes(details *ServerDB) {
	userController := NewUserController(details.Db)

	orders := details.Server.Group("/products/order")
	{
		orders.POST("/:productID", userController.OrderProduct)
		orders.GET("/get/:order-ID", userController.GetOrder)
		orders.GET("/get", userController.GetOrdersWithUsername)
		orders.PUT("/receive/:order-id", userController.ReceiveOrder)
		orders.POST("/cart", userController.OrderAllCartItems)
	}
}

func configureCartCollection(details *ServerDB) error {
	collection := details.Db.Collection("cart")

	_, err := collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "product_name", Value: 1}},
		Options: options.Index().SetName("product_name_index").SetUnique(true),
	})
	return err
}

func initializeCartRoutes(details *ServerDB) {
	configureCartCollection(details)

	userController := NewUserController(details.Db)
	cart := details.Server.Group("/user/cart")
	{
		cart.POST("/add", userController.AddToCart)
		cart.DELETE("/remove/:cart-id", userController.RemoveFromCart)
		cart.PUT("/update", userController.UpdateCartQuantity)
		cart.GET("/getall", userController.GetUserCartItems)
	}
}
