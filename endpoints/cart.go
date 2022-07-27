package endpoints

import (
	"github.com/Emmrys-Jay/ecommerce-api/controller"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeCartEndpoints(db *mongo.Database, e *gin.Engine, mdw gin.HandlerFunc) {
	usercontroller := controller.NewUserController(db)

	cart := e.Group("/user/cart", mdw)
	{
		cart.POST("/add", usercontroller.AddToCart)
		cart.DELETE("/remove/:cart-id", usercontroller.RemoveFromCart)
		cart.PUT("/decrement/:cart-id", usercontroller.DecrementCartQuantity)
		cart.GET("/getall", usercontroller.GetUserCartItems)
	}
}
