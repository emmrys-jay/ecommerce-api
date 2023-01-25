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
		user.POST("/create", userController.CreateUser)
		user.POST("/login", userController.LoginUser)
		// user.POST("/logout", userController.LogoutUser)
		user.GET("/get", mdw, userController.GetUser)
		user.PUT("/password", mdw, userController.ChangePassword)
		user.PUT("/update", mdw, userController.UpdateUserFlexible)
		user.PUT("/location/add", mdw, userController.AddLocation)
	}
}
