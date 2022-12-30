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
		admin.GET("/cart/:cart-id", adminController.GetCartItem)
		admin.GET("/cart/get_all", adminController.GetAllCartItems)
		admin.DELETE("/cart/:id", adminController.DeleteCartItem)
		admin.DELETE("/cart/delete_all/:username", adminController.DeleteAllUserCartItems)
		admin.DELETE("/cart/delete_all", adminController.DeleteAllCartItems)

		admin.GET("/orders/get_all", adminController.GetAllOrders)
		admin.PATCH("/deliver/:order-id", adminController.DeliverOrder)
		admin.DELETE("/orders/:id", adminController.DeleteOrder)
		admin.DELETE("/orders/delete_all/:user-id", adminController.DeleteAllOrdersWithUserID)
		admin.DELETE("/orders/delete_all", adminController.DeleteAllOrders)

		admin.POST("/products/add_one", adminController.AddOneProduct)
		admin.POST("/products", adminController.AddProducts)
		admin.DELETE("/products", adminController.DeleteProducts)
		admin.DELETE("/products/delete_all", adminController.DeleteAllProducts)
		admin.PATCH("/products/:id", adminController.UpdateProduct)
		//products.GET("/categories", getAllCategories)

		admin.GET("/user/:user-id", adminController.GetUser)
		admin.GET("/user/get_all", adminController.GetAllUsers)
		admin.PATCH("/user", adminController.UpdateUserFlexible)
		admin.DELETE("/user/:user-id", adminController.DeleteUser)
		admin.DELETE("/user/delete_all", adminController.DeleteAllUsers)
	}
}
