package handlers

import (
	"github.com/Emmrys-Jay/ecommerce-api/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// Server stores the gin engine and mongo database to be used as method receivers
type Server struct {
	DB     *mongo.Database
	Server *gin.Engine
}

// StartServer defines all endpoints and associates them with their handlers
func StartServer(server *Server, addr string) {

	products := server.Server.Group("/products", middleware.AuthorizeJWT())
	{
		products.POST("/addone", server.AddOneProduct)
		products.POST("/add", server.AddProducts)
		products.GET("/find", server.FindProducts)
		products.DELETE("/delete", server.DeleteProduct)
		products.DELETE("/deleteall", server.DeleteAllProducts)
		products.PUT("/update", server.UpdateProduct)
		products.PUT("/addreview", server.AddReview)
		//products.GET("/categories", getAllCategories)
	}

	users := server.Server.Group("/users")
	{
		users.POST("/create", server.CreateUser)
		users.POST("/login", server.LoginUser)
		users.GET("/getAll", middleware.AuthorizeJWT(), server.GetAllUsers)
		users.GET("/get", middleware.AuthorizeJWT(), server.GetUser)
		users.PUT("/password", middleware.AuthorizeJWT(), server.ChangePassword)
		users.PUT("/update", middleware.AuthorizeJWT(), server.UpdateUserFlexible)
		users.PUT("/location/add", middleware.AuthorizeJWT(), server.AddLocation)
		users.DELETE("/delete", middleware.AuthorizeJWT(), server.DeleteUser)
		users.DELETE("/deleteAll", middleware.AuthorizeJWT(), server.DeleteAllUsers)
	}

	cart := server.Server.Group("/user/cart", middleware.AuthorizeJWT())
	{
		cart.POST("/add", server.AddToCart)
		cart.DELETE("/delete", server.RemoveFromCart)
		cart.PUT("/decrement", server.DecrementCartQuantity)
		cart.GET("/get", server.GetCartItem)
		cart.GET("/getAll", server.GetUserCartItems)
		cart.DELETE("/deleteAll", server.DeleteAllCartItems)
	}

	server.Server.Run(addr)
}

// errorResponse returns verbose error response
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
