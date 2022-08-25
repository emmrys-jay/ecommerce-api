package endpoints

import (
	"github.com/Emmrys-Jay/ecommerce-api/controller"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeProductEndpoints(db *mongo.Database, e *gin.Engine, mdw gin.HandlerFunc) {
	userController := controller.NewUserController(db)

	products := e.Group("/products")
	{
		products.GET("/get/:category", userController.GetProductsByCategory)
		products.GET("/find", userController.FindProducts)
		products.GET("/findone/:productID", userController.FindOneProduct)
		// products.GET("/find/recent", userController.FindProductsWithTime)
		// products.GET("/find/star", userController.FindProductsBasedOnReviews)
		products.PUT("/:productID/addreview", mdw, userController.AddReview)
		// products.GET("/categories", getAllCategories)
	}
}
