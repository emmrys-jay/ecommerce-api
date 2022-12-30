package endpoints

import (
	"github.com/Emmrys-Jay/ecommerce-api/controller"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeCartEndpoints(db *mongo.Database, e *gin.Engine, mdw gin.HandlerFunc) {
	usercontroller := controller.NewUserController(db)

	cart := e.Group("/user", mdw)
	{
		cart.POST("/cart", usercontroller.AddToCart)
		cart.DELETE("/cart/:cart-id", usercontroller.RemoveFromCart)
		cart.PUT("/cart", usercontroller.UpdateCartQuantity)
		cart.GET("/cart/get_all", usercontroller.GetUserCartItems)
	}
}
