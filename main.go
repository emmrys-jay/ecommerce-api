package main

import (
	"github.com/Emmrys-Jay/ecommerce-api/db"
	api "github.com/Emmrys-Jay/ecommerce-api/handlers"
	"github.com/gin-gonic/gin"
)

var server = api.Server{}

func main() {
	db.ConfigDB()
	server = handlers.Server{
		DB:     db.DB,
		Server: gin.Default(),
	}

	api.StartServer(&server, ":8080")
}
