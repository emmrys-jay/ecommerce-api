package endpoints

import (
	"github.com/Emmrys-Jay/ecommerce-api/controller"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeUserEndpoints(db *mongo.Database, e *gin.Engine, mdw gin.HandlerFunc) {
	userController := controller.NewUserController(db)

	users := e.Group("/users")
	{
		users.POST("/create", userController.CreateUser)
		users.POST("/login", userController.LoginUser)
		users.GET("/get/:username", mdw, userController.GetUser)
		users.PUT("/password", mdw, userController.ChangePassword)
		users.PUT("/update", mdw, userController.UpdateUserFlexible)
		users.PUT("/location/add", mdw, userController.AddLocation)
	}
}
