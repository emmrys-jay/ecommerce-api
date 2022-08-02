package main

import (
	"log"
	"os"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/endpoints"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/middleware"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	"github.com/Emmrys-Jay/ecommerce-api/util"
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

	adminPassword, err := util.HashPassword("ADMIN" + adminPassword)
	if err != nil {
		log.Fatalln("could not hash password")
	}

	_, err = repository.GetUser(db.GetCollection(database, "users"), adminUsername)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err := createAdminUserMain(database, adminUsername, adminPassword)
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
	//gin.SetMode(gin.ReleaseMode)

	adminMdw := middleware.AuthorizeAdmin(adminUsername)

	userMdw := middleware.AuthorizeJWT()

	// Initialize endpoints
	endpoints.InitializeAdminEndpoints(database, server, adminMdw)
	endpoints.InitializeCartEndpoints(database, server, userMdw)
	endpoints.InitializeUserEndpoints(database, server, userMdw)
	endpoints.InitializeProductEndpoints(database, server, userMdw)
	endpoints.InitializeOrdersEndpoints(database, server, userMdw)

	err = server.Run()
	if err != nil {
		log.Fatalln("could not start server")
	}

}

func createAdminUserMain(database *mongo.Database, adminUsername, adminPassword string) error {
	admin := entity.User{
		ID:             primitive.NewObjectID().String(),
		Username:       adminUsername,
		PasswordSalt:   "ADMIN",
		HashedPassword: adminPassword,
		Fullname:       "ADMIN",
		Email:          "ADMIN",
		CreatedAt:      time.Now(),
	}

	_, err := repository.CreateUser(db.GetCollection(database, "users"), admin)
	if err != nil {
		return err
	}

	return nil
}
