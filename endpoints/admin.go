package endpoints

import (
	"github.com/Emmrys-Jay/ecommerce-api/controller"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeAdminEndpoints(db *mongo.Database, e *gin.Engine, mdw gin.HandlerFunc) {
	userController := controller.NewUserController(db)
	adminController := controller.NewAdminController(userController)

	admin := e.Group("/admin/user", mdw)
	{
		admin.GET("/cart/get", adminController.GetCartItem)
		admin.DELETE("/cart/deleteAll", adminController.DeleteAllCartItems)

		admin.GET("/orders/getAll", adminController.GetAllOrders)

		admin.DELETE("/product/delete", adminController.DeleteProduct)
		admin.DELETE("/product/deleteAll", adminController.DeleteAllProducts)
		admin.PUT("/product/update", adminController.UpdateProduct)

		admin.GET("/get", adminController.GetAllUsers)
		admin.DELETE("/delete", adminController.DeleteUser)
		admin.DELETE("/deleteAll", adminController.DeleteAllUsers)
		//products.GET("/categories", getAllCategories)
	}
}
