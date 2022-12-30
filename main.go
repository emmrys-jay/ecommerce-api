package main

import (
	"log"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/endpoints"
	"github.com/Emmrys-Jay/ecommerce-api/middleware"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load("load.env")

	database := db.ConfigDB()

	// Create admin user in database
	adminUsername, err := repository.CreateAdminUser(database)
	if err != nil {
		log.Fatalln(err)
	}

	// Get middlewares to verify admin and users
	adminMdw := middleware.AuthorizeAdmin(adminUsername)
	userMdw := middleware.AuthorizeJWT()

	// Create server
	server := gin.New()

	// Setup routes
	endpoints.SetupRoutes(database, server, adminMdw, userMdw)

	log.Fatalln(server.Run())
}
