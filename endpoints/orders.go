package endpoints

import (
	"github.com/Emmrys-Jay/ecommerce-api/controller"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeOrdersEndpoints(db *mongo.Database, e *gin.Engine, mdw gin.HandlerFunc) {
	userController := controller.NewUserController(db)

	orders := e.Group("/products/order", mdw)
	{
		orders.POST("/", userController.OrderProduct)
		orders.GET("/get", userController.GetOrdersWithUsername)

	}
}
