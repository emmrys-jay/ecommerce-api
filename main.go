package main

import (
	"github.com/Emmrys-Jay/ecommerce-api/db"
	handlers "github.com/Emmrys-Jay/ecommerce-api/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var server = handlers.Server{}

func main() {

	_ = godotenv.Load("./token/load.env")

	db.ConfigDB()
	server = handlers.Server{
		DB:     db.DB,
		Server: gin.Default(),
	}

	handlers.StartServer(&server, ":8080")
}
