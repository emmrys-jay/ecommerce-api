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

func initializeUserRoutes(ed *ServerDB) {
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
	initializeUserRoutes(&details)

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
	initializeUserRoutes(&details)

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

	expectedBody := UserResponse{}

	if len(triggers) > 1 {
		if triggers[1] == "unauthorised" {
			assert.Equal(t, 401, recorder.Code)
			assert.NotZero(t, recorder.Body.String)
		}
	} else {
		err = json.Unmarshal(recorder.Body.Bytes(), &expectedBody)
		assert.NoError(t, err)
		assert.Equal(t, 200, recorder.Code, expectedBody)
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
	return expectedBody.Token
}

func TestLoginUser(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}
	initializeUserRoutes(&details)

	username := "Harry"
	createUserTest(t, &details, username)
	token := loginUserTest(t, &details, username)

	assert.NotZero(t, token)

	assert.NoError(t, deleteUsers(details.Db))
	assert.NoError(t, dropDatabase(details.Db))
}

func getUserTest(t *testing.T, details *ServerDB, user UserResponse) GetUserResponse {
	req, _ := http.NewRequest("GET", "/user/get", nil)
	req.Header.Add("Authorization", "Bearer "+user.Token)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)

	var gUser = GetUserResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &gUser)
	assert.NoError(t, err)

	assert.Equal(t, 200, recorder.Code)

	assert.Equal(t, user.Username, gUser.Username)
	assert.Equal(t, user.Fullname, gUser.Fullname)
	assert.Equal(t, user.Email, gUser.Email)
	assert.Equal(t, user.ID, gUser.ID)
	assert.Equal(t, user.MobileNumber, gUser.MobileNumber)

	return gUser
}

func TestGetUser(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}
	initializeUserRoutes(&details)

	username := "Harry"
	user := createUserTest(t, &details, username)

	_ = getUserTest(t, &details, user)

	assert.NoError(t, deleteUsers(details.Db))
	assert.NoError(t, dropDatabase(details.Db))
}

func TestChangePassword(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}
	initializeUserRoutes(&details)

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
	assert.Equal(t, 200, recorder.Code)
	assert.NotEqual(t, "404 page not found", recorder.Body.String())

	token := loginUserTest(t, &details, username, "", "unauthorised")
	assert.Empty(t, token)

	token = loginUserTest(t, &details, username, "reversed")
	assert.NotEmpty(t, token)

	assert.NoError(t, deleteUsers(details.Db))
	assert.NoError(t, dropDatabase(details.Db))
}

func updateUserTest(t *testing.T, details *ServerDB, token string, updateDetail string, user GetUserResponse) {
	uReq := UpdateUserRequest{}

	switch updateDetail {
	case "username":
		uReq = UpdateUserRequest{
			Detail: updateDetail,
			Update: "NewUsername",
		}
	case "profile_picture":
		uReq = UpdateUserRequest{
			Detail: updateDetail,
			Update: "NewProfilePicture",
		}
	case "email":
		uReq = UpdateUserRequest{
			Detail: updateDetail,
			Update: "NewEmail",
		}
	case "mobile_number":
		uReq = UpdateUserRequest{
			Detail: updateDetail,
			Update: "NewMobileNumber",
		}
	case "default_payment_method":
		uReq = UpdateUserRequest{
			Detail: updateDetail,
			Update: "NewMobileNumber",
		}
	}

	cpReqJson, _ := json.Marshal(uReq)
	req, _ := http.NewRequest("PUT", "/user/update", bytes.NewBuffer(cpReqJson))
	req.Header.Add("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code)
	assert.NotEqual(t, "404 page not found", recorder.Body.String())

	req, _ = http.NewRequest("GET", "/user/get", nil)
	req.Header.Add("Authorization", "Bearer "+token)

	recorder = httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code)
	assert.NotEqual(t, "404 page not found", recorder.Body.String())

	var updatedUser = GetUserResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &updatedUser)
	assert.NoError(t, err)

	switch updateDetail {
	case "username":
		assert.NotEqual(t, user.Username, updatedUser.Username)
	case "profile_picture":
		assert.NotEqual(t, user.ProfilePicture, updatedUser.ProfilePicture)
	case "email":
		assert.NotEqual(t, user.Email, updatedUser.Email)
	case "mobile_number":
		assert.NotEqual(t, user.MobileNumber, updatedUser.MobileNumber)
	case "default_payment_method":
		assert.NotEqual(t, user.DefaultPaymentMethod, updatedUser.DefaultPaymentMethod)
	}

	assert.True(t, updatedUser.LastUpdated.After(user.LastUpdated))
}

func TestUpdateUserFlexible(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}
	initializeUserRoutes(&details)

	username := "Harry"
	user := createUserTest(t, &details, username)

	gUser := getUserTest(t, &details, user)

	// test updated profile picture
	updateUserTest(t, &details, user.Token, "profile_picture", gUser)

	// test updated default payment method
	updateUserTest(t, &details, user.Token, "default_payment_method", gUser)

	// test updated email
	updateUserTest(t, &details, user.Token, "email", gUser)

	// test updated mobile number
	updateUserTest(t, &details, user.Token, "mobile_number", gUser)

	// test updated username
	updateUserTest(t, &details, user.Token, "username", gUser)

	assert.NoError(t, deleteUsers(details.Db))
	assert.NoError(t, dropDatabase(details.Db))
}

func addLocationTest(t *testing.T, details *ServerDB, user UserResponse, location entity.Location, triggers ...string) {
	lreq := AddLocationRequest{
		Location: location,
	}

	lreqJson, err := json.Marshal(lreq)
	assert.NoError(t, err)

	req, _ := http.NewRequest("PUT", "/user/location/add", bytes.NewBuffer(lreqJson))
	req.Header.Add("Authorization", "Bearer "+user.Token)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code)

	updatedUser := getUserTest(t, details, user)

	if len(triggers) > 0 {
		if triggers[0] == "second_location" {
			assert.NotEqual(t, location, updatedUser.DefaultDeliveryLocation)
			assert.NotEqual(t, location, updatedUser.RegisteredLocations[0])
			assert.Equal(t, location, updatedUser.RegisteredLocations[1])
			assert.Equal(t, location.HouseNumber, updatedUser.RegisteredLocations[1].HouseNumber)
			assert.Equal(t, location.PhoneNo, updatedUser.RegisteredLocations[1].PhoneNo)
			assert.Equal(t, location.Street, updatedUser.RegisteredLocations[1].Street)
			assert.Equal(t, location.CityOrTown, updatedUser.RegisteredLocations[1].CityOrTown)
			assert.Equal(t, location.State, updatedUser.RegisteredLocations[1].State)
			assert.Equal(t, location.Country, updatedUser.RegisteredLocations[1].Country)
			assert.Equal(t, location.ZipCode, updatedUser.RegisteredLocations[1].ZipCode)
		}
	} else {
		assert.NotEmpty(t, updatedUser.RegisteredLocations)
		assert.Equal(t, location, updatedUser.RegisteredLocations[0])
		assert.Equal(t, location, updatedUser.DefaultDeliveryLocation)
		assert.Equal(t, location.HouseNumber, updatedUser.RegisteredLocations[0].HouseNumber)
		assert.Equal(t, location.PhoneNo, updatedUser.RegisteredLocations[0].PhoneNo)
		assert.Equal(t, location.Street, updatedUser.RegisteredLocations[0].Street)
		assert.Equal(t, location.CityOrTown, updatedUser.RegisteredLocations[0].CityOrTown)
		assert.Equal(t, location.State, updatedUser.RegisteredLocations[0].State)
		assert.Equal(t, location.Country, updatedUser.RegisteredLocations[0].Country)
		assert.Equal(t, location.ZipCode, updatedUser.RegisteredLocations[0].ZipCode)
	}
}

func TestAddLocation(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}
	initializeUserRoutes(&details)

	username := "Harry"
	user := createUserTest(t, &details, username)

	location := entity.Location{
		HouseNumber: "88a",
		PhoneNo:     "+2346",
		Street:      "Adelaide",
		CityOrTown:  "Right City",
		State:       "Rely land",
		Country:     "Moa",
		ZipCode:     "2377809",
	}

	// Test for first added location being also assigned default location
	addLocationTest(t, &details, user, location)

	location = entity.Location{
		HouseNumber: "50",
		PhoneNo:     "+2346446",
		Street:      "Mawuli",
		CityOrTown:  "Left City",
		State:       "Sroboth",
		Country:     "Land",
	}

	// Test for a second location
	addLocationTest(t, &details, user, location, "second_location")

	assert.NoError(t, deleteUsers(details.Db))
	assert.NoError(t, dropDatabase(details.Db))
}
