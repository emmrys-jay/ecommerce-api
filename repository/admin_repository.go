package repository

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateAdminUser creates an admin user based on the specified environment variables
func CreateAdminUser(database *mongo.Database) (adminUsername string, retErr error) {
	adminUsername = os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		return "", errors.New("admin username not specified")
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		return "", errors.New("admin password not specified")
	}

	adminPassword, err := util.HashPassword("ADMIN" + adminPassword)
	if err != nil {
		return "", fmt.Errorf("error hashing admin password: %v", err)
	}

	userCollection := db.GetCollection(database, "users")
	_, err = GetUser(userCollection, "", adminUsername)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			admin := entity.User{
				ID:           primitive.NewObjectID().String()[10:34],
				Username:     adminUsername,
				PasswordSalt: "ADMIN",
				Password:     adminPassword,
				Fullname:     "ADMIN",
				Email:        "ADMIN",
				CreatedAt:    time.Now(),
			}

			_, err := CreateUser(db.GetCollection(database, "users"), admin)
			if err != nil {
				return "", fmt.Errorf("error creating admin user: %v", err)
			}
		} else {
			return "", fmt.Errorf("error checking for admin user: %v", err)
		}
	}

	return adminUsername, nil
}
