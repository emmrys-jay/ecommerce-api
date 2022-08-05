package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/endpoints"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

var server = gin.Default()
var Database *mongo.Database

func hashPassword(password string) string {
	hp, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalln("Could not hash password")
	}

	return string(hp)
}

func initializeUserRoutes() {
	mdw := middleware.AuthorizeJWT()
	endpoints.InitializeUserEndpoints(Database, server, mdw)
}

func connectDB() *mongo.Database {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalln("could not connect to server: ", err)
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalln("could not ping server: ", err)
	}

	return client.Database("ecommerce_test")
}

func TestCreateUser(t *testing.T) {
	Database = connectDB()
	godotenv.Load("../load.env")

	user := entity.User{
		Username:       "Ken",
		Fullname:       "Thompson Kenny",
		HashedPassword: hashPassword("kenny101"),
		Email:          "ken@email.com",
	}

	userJson, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", "/user/create", bytes.NewBuffer(userJson))
	if err != nil {
		log.Fatalln("Could not create request")
	}

	recorder := httptest.NewRecorder()

	initializeUserRoutes()
	server.ServeHTTP(recorder, req)

	var expectedBody = struct {
		ID             string    `json:"_id"`
		Username       string    `json:"username"`
		Fullname       string    `json:"fullname"`
		Email          string    `json:"email"`
		EmailIsVerfied bool      `json:"email_is_verified"`
		MobileNumber   string    `json:"mobile_number,omitempty"`
		CreatedAt      time.Time `json:"created_at"`
	}{}

	err = json.Unmarshal(recorder.Body.Bytes(), &expectedBody)
	assert.Nil(t, err)
	assert.Equal(t, "", expectedBody.MobileNumber)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.NotZero(t, expectedBody.ID)
	assert.Equal(t, user.Username, expectedBody.Username)
	assert.Equal(t, user.Fullname, expectedBody.Fullname)
	assert.Equal(t, user.Email, expectedBody.Email)
	assert.Equal(t, false, expectedBody.EmailIsVerfied)
	assert.True(t, expectedBody.CreatedAt.Before(time.Now()))
}
