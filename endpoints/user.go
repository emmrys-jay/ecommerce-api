package endpoints

import (
	"github.com/Emmrys-Jay/ecommerce-api/controller"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeUserEndpoints(db *mongo.Database, e *gin.Engine, mdw gin.HandlerFunc) {
	userController := controller.NewUserController(db)

	user := e.Group("/user")
	{
		user.GET("", mdw, userController.GetUser)
		user.PATCH("", mdw, userController.UpdateUserFlexible)
		user.POST("/signup", userController.CreateUser)
		user.POST("/login", userController.LoginUser)
		user.PATCH("/password", mdw, userController.ChangePassword)
		user.PATCH("/location", mdw, userController.AddLocation)
	}
}
