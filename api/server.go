package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	DB     *mongo.Database
	Server *gin.Engine
}

func StartServer(server *Server, addr string) {

	products := server.Server.Group("/products")
	{
		products.POST("/addone", server.AddOneProduct)
		products.POST("/add", server.AddProducts)
		products.GET("/find", server.FindProducts)
		products.DELETE("/delete", server.DeleteProduct)
		products.DELETE("/deleteall", server.DeleteAllProducts)
		products.PUT("/update", server.UpdateProduct)
		products.POST("/addreview", server.AddReview)
		//products.GET("/categories", getAllCategories)
	}

	users := server.Server.Group("/users")
	{
		users.POST("/create", server.CreateUser)
	}

	server.Server.Run(addr)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
