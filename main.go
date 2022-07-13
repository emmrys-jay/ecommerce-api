package main

import (
	"github.com/Emmrys-Jay/ecommerce-api/db"
	handlers "github.com/Emmrys-Jay/ecommerce-api/handlers"
	"github.com/gin-gonic/gin"
)

var server = handlers.Server{}

func main() {
	db.ConfigDB()
	server = handlers.Server{
		DB:     db.DB,
		Server: gin.Default(),
	}

	handlers.StartServer(&server, ":8080")
}
