package main

import (
	"github.com/Emmrys-Jay/ecommerce-api/api"
	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/gin-gonic/gin"
)

var server = api.Server{}

func main() {
	db.ConfigDB()
	server = api.Server{
		DB:     db.DB,
		Server: gin.Default(),
	}

	api.StartServer(&server, ":8080")
}
