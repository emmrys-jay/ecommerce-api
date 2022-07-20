package main

import (
	"log"
	"os"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/endpoints"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/middleware"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {

	_ = godotenv.Load("load.env")

	database := db.ConfigDB()

	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		log.Fatalln("admin username not specified")
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatalln("admin password not specified")
	}

	_, err := repository.GetUser(db.GetCollection(database, "user"), adminUsername)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			admin := entity.User{
				ID:             primitive.NewObjectID(),
				Username:       adminUsername,
				PasswordSalt:   "ADMIN",
				HashedPassword: adminPassword,
				Fullname:       "ADMIN",
				Email:          "ADMIN",
			}

			_, err := repository.CreateUser(db.GetCollection(database, "user"), admin)
			if err != nil {
				log.Fatalln("could not create admin user")
			}
		} else {
			log.Fatalln("something went wrong")
		}
	}

	// server := gin.New()
	// server.Use()

	server := gin.Default()

	adminMdw := middleware.AuthorizeAdmin(adminUsername)

	userMdw := middleware.AuthorizeJWT()

	// Initialize endpoints
	endpoints.InitializeAdminEndpoints(database, server, adminMdw)
	endpoints.InitializeCartEndpoints(database, server, userMdw)
	endpoints.InitializeUserEndpoints(database, server, userMdw)
	endpoints.InitializeProductEndpoints(database, server, userMdw)

	err = server.Run()
	if err != nil {
		log.Fatalln("could not start server")
	}

}
