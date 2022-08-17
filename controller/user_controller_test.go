package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

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

func deleteUsers(database *mongo.Database) error {
	c := database.Collection("users")
	_, err := c.DeleteMany(context.Background(), bson.M{})
	return err
}

func initializeCreateUserRoutes(ed *ServerDB) {
	userController := NewUserController(ed.Db)

	user := ed.Server.Group("/user")
	{
		user.POST("/create", userController.CreateUser)
		user.POST("/login", userController.LoginUser)
		// user.POST("/logout", userController.LogoutUser)
		user.GET("/get", userController.GetUser)
		user.PUT("/password", userController.ChangePassword)
		user.PUT("/update", userController.UpdateUserFlexible)
		user.PUT("/location/add", userController.AddLocation)
	}
}

func dropDatabase(d *mongo.Database) error {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)

	err := d.Drop(ctx)
	err = d.Client().Disconnect(ctx)

	return err
}

func createUserTest(t *testing.T, details *ServerDB, username string) UserResponse {
	initializeCreateUserRoutes(details)
	godotenv.Load("../load.env")
	user := entity.User{
		Username:       username,
		Fullname:       "Thompson " + username,
		HashedPassword: "101" + username,
		Email:          username + "@email.com",
	}

	userJson, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", "/user/create", bytes.NewBuffer(userJson))
	if err != nil {
		log.Fatalln("Could not create request")
	}

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)

	var expectedBody = UserResponse{}

	err = json.Unmarshal(recorder.Body.Bytes(), &expectedBody)
	assert.NoError(t, err)
	assert.Equal(t, "", expectedBody.MobileNumber)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.NotZero(t, expectedBody.ID)
	assert.NotZero(t, expectedBody.Token)
	assert.Equal(t, user.Username, expectedBody.Username)
	assert.Equal(t, user.Fullname, expectedBody.Fullname)
	assert.Equal(t, user.Email, expectedBody.Email)
	assert.Equal(t, false, expectedBody.EmailIsVerfied)
	assert.True(t, expectedBody.CreatedAt.Before(time.Now()))

	return expectedBody
}

func TestCreateUser(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}

	createUserTest(t, &details, "Harry")
	assert.NoError(t, deleteUsers(details.Db))
	assert.NoError(t, dropDatabase(details.Db))
}

func configureDbCollections(db *mongo.Database) error {
	collection := db.Collection("users")

	_, err := collection.Indexes().CreateMany(context.Background(),
		[]mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "username", Value: 1}},
				Options: options.Index().SetName("username_index").SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetName("email_index").SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "mobile_number", Value: 1}},
				Options: options.Index().SetName("mobile_number_index").SetUnique(true),
			},
		})

	return err
}

func TestUniqueUserFields(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}

	initializeCreateUserRoutes(&details)
	configureDbCollections(details.Db)
	godotenv.Load("../load.env")

	username := "Harry"
	user := entity.User{
		Username:       username,
		Fullname:       "Thompson " + username,
		HashedPassword: "101" + username,
		Email:          username + "@email.com",
		MobileNumber:   "0700",
	}

	userJson, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", "/user/create", bytes.NewBuffer(userJson))
	if err != nil {
		log.Fatalln("Could not create request")
	}

	recorder1 := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder1, req)
	assert.Equal(t, 200, recorder1.Code)

	recorder2 := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder2, req)
	assert.Equal(t, 400, recorder2.Code)

	assert.NotEqual(t, recorder1.Body.Bytes(), recorder2.Body.Bytes())

	username = "Tom"
	user = entity.User{
		Username:       username,
		Fullname:       "Thompson " + username,
		HashedPassword: "101" + username,
		Email:          username + "@email.com",
		MobileNumber:   "0800",
	}

	userJson, _ = json.Marshal(user)
	req, err = http.NewRequest("POST", "/user/create", bytes.NewBuffer(userJson))
	if err != nil {
		log.Fatalln("Could not create request")
	}

	recorder1 = httptest.NewRecorder()
	details.Server.ServeHTTP(recorder1, req)
	assert.Equal(t, 200, recorder1.Code)

	assert.NoError(t, deleteUsers(details.Db))
	assert.NoError(t, dropDatabase(details.Db))
}

func loginUserTest(t *testing.T, details *ServerDB, username string, triggers ...string) string {
	lreq := LoginUserRequest{
		Username: username,
		Password: "101" + username,
	}

	if len(triggers) > 0 {
		if triggers[0] == "reversed" {
			lreq.Password = username + "101"
		}
	}

	lreqJson, _ := json.Marshal(lreq)
	req, err := http.NewRequest("POST", "/user/login", bytes.NewBuffer(lreqJson))
	if err != nil {
		log.Fatalln("Could not create request")
	}

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code)

	expectedBody := UserResponse{}

	if len(triggers) > 1 {
		if triggers[1] == "unauthorised" {
			require.Equal(t, 401, recorder.Body)
			require.NotEqual(t, "404 page not found", recorder.Body.String)
		} else {
			err = json.Unmarshal(recorder.Body.Bytes(), &expectedBody)
			assert.NoError(t, err)
			assert.Equal(t, "", expectedBody.MobileNumber)

			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.NotZero(t, expectedBody.ID)
			assert.NotZero(t, expectedBody.Token)
			assert.Equal(t, expectedBody.Username, expectedBody.Username)
			assert.Equal(t, expectedBody.Fullname, expectedBody.Fullname)
			assert.Equal(t, expectedBody.Email, expectedBody.Email)
			assert.Equal(t, false, expectedBody.EmailIsVerfied)
			assert.True(t, expectedBody.CreatedAt.Before(time.Now()))
		}
	}
	return expectedBody.Token
}

func TestLoginUser(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}

	username := "Harry"
	createUserTest(t, &details, username)
	token := loginUserTest(t, &details, username)

	assert.NotZero(t, token)

	assert.NoError(t, deleteUsers(details.Db))
	assert.NoError(t, dropDatabase(details.Db))
}

func TestGetUser(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}

	username := "Harry"
	user := createUserTest(t, &details, username)

	req, _ := http.NewRequest("GET", "/user/get", nil)
	req.Header.Add("Authorization", "Bearer "+user.Token)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)

	var gUser = GetUserResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &gUser)
	require.NoError(t, err)

	require.Equal(t, 200, recorder.Code)

	require.Equal(t, user.Username, gUser.Username)
	require.Equal(t, user.Fullname, gUser.Fullname)
	require.Equal(t, user.Email, gUser.Email)
	require.Equal(t, user.ID, gUser.ID)
	require.Equal(t, user.MobileNumber, gUser.MobileNumber)

	require.NoError(t, deleteUsers(details.Db))
	require.NoError(t, dropDatabase(details.Db))
}

func TestChangePassword(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}

	username := "Harry"
	user := createUserTest(t, &details, username)

	cpReq := ChangePasswordRequest{
		Password:    "101" + username,
		NewPassword: username + "101",
	}

	cpReqJson, _ := json.Marshal(cpReq)
	req, _ := http.NewRequest("PUT", "/user/password", bytes.NewBuffer(cpReqJson))
	req.Header.Add("Authorization", "Bearer "+user.Token)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)
	require.NotEqual(t, "404 page not found", recorder.Body.String())

	require.NoError(t, deleteUsers(details.Db))
	require.NoError(t, dropDatabase(details.Db))

}
