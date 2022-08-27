package endpoints

import (
	"github.com/Emmrys-Jay/ecommerce-api/controller"
	admin "github.com/Emmrys-Jay/ecommerce-api/controller/admin"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitializeAdminEndpoints(db *mongo.Database, e *gin.Engine, mdw gin.HandlerFunc) {
	userController := controller.NewUserController(db)
	adminController := admin.NewAdminController(userController)

	admin := e.Group("/admin", mdw)
	{
		admin.GET("/cart/get/:cart-id", adminController.GetCartItem)
		admin.GET("/cart/getall", adminController.GetAllCartItems)
		admin.DELETE("/cart/delete/:id", adminController.DeleteCartItem)
		admin.DELETE("/cart/deleteall/:username", adminController.DeleteAllUserCartItems)
		admin.DELETE("/cart/deleteall", adminController.DeleteAllCartItems)

		admin.GET("/orders/getall", adminController.GetAllOrders)
		admin.PUT("/deliver/:order-id", adminController.DeliverOrder)
		admin.DELETE("/orders/delete/:id", adminController.DeleteOrder)
		admin.DELETE("/orders/deleteall/:username", adminController.DeleteAllOrdersWithUsername)
		admin.DELETE("/orders/deleteall", adminController.DeleteAllOrders)

		admin.POST("/products/addone", adminController.AddOneProduct)
		admin.POST("/products/add", adminController.AddProducts)
		admin.DELETE("/products/delete", adminController.DeleteProduct)
		admin.DELETE("/products/deleteall", adminController.DeleteAllProducts)
		admin.PUT("/products/update", adminController.UpdateProduct)

		admin.GET("/user/get/:username", adminController.GetUser)
		admin.GET("/user/getAll", adminController.GetAllUsers)
		admin.PUT("/user/update", adminController.UpdateUserFlexible)
		admin.PUT("/user/location/add", mdw, adminController.AddLocation)
		admin.DELETE("/user/delete", adminController.DeleteUser)
		admin.DELETE("/user/deleteAll", adminController.DeleteAllUsers)
		//products.GET("/categories", getAllCategories)
	}
}
