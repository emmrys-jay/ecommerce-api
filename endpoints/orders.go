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
		orders.POST("/:productID", userController.OrderProduct)
		orders.GET("/:order-ID", userController.GetOrder)
		orders.GET("/orders", userController.GetOrdersWithUsername)
		orders.PATCH("/receive/:order-id", userController.ReceiveOrder)
		orders.POST("/cart", userController.OrderAllCartItems)
	}
}
