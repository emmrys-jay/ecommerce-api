package endpoints

import (
	"github.com/Emmrys-Jay/ecommerce-api/controller"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeProductEndpoints(db *mongo.Database, e *gin.Engine, mdw gin.HandlerFunc) {
	userController := controller.NewUserController(db)

	products := e.Group("/products", mdw)
	{
		products.POST("/addone", userController.AddOneProduct)
		products.POST("/add", userController.AddProducts)
		products.GET("/find", userController.FindProducts)
		products.PUT("/addreview", userController.AddReview)
		//products.GET("/categories", getAllCategories)
	}
}
