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
		products.GET("/:category", userController.GetProductsByCategory)
		products.GET("/find", userController.FindProducts)
		products.GET("/find_one/:productID", userController.FindOneProduct)
		// products.GET("/find/recent", userController.FindProductsWithTime)
		// products.GET("/find/reviews", userController.FindProductsBasedOnReviews)
		products.PATCH("/:productID/add_review", mdw, userController.AddReview)
		// products.GET("/categories", getAllCategories)
	}
}
