// user_test

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
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func deleteRecords(database *mongo.Database, collection string) error {
	c := database.Collection(collection)
	_, err := c.DeleteMany(context.Background(), bson.M{})
	return err
}

func dropDatabase(d *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := d.Drop(ctx)
	err = d.Client().Disconnect(ctx)

	return err
}

func createUserTest(t *testing.T, details *ServerDB, username string) entity.UserResponse {
	godotenv.Load("../load.env")
	user := entity.User{
		Username: username,
		Fullname: "Thompson " + username,
		Password: "101" + username,
		Email:    username + "@email.com",
	}

	userJson, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", "/user/create", bytes.NewBuffer(userJson))
	if err != nil {
		log.Fatalln("Could not create request")
	}

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)

	var expectedBody = entity.UserResponse{}

	err = json.Unmarshal(recorder.Body.Bytes(), &expectedBody)
	require.NoError(t, err)
	require.Equal(t, "", expectedBody.MobileNumber)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotZero(t, expectedBody.ID)
	require.NotZero(t, expectedBody.Token)
	require.Equal(t, user.Username, expectedBody.Username)
	require.Equal(t, user.Fullname, expectedBody.Fullname)
	require.Equal(t, user.Email, expectedBody.Email)
	require.Equal(t, false, expectedBody.EmailIsVerfied)
	require.True(t, expectedBody.CreatedAt.Before(time.Now()))

	return expectedBody
}

func TestCreateUser(t *testing.T) {
	details := NewServerDB()

	initializeUserRoutes(details)

	createUserTest(t, details, "Harry")
	require.NoError(t, deleteRecords(details.Db, "users"))
	require.NoError(t, dropDatabase(details.Db))
}

func configureUserCollection(db *mongo.Database) error {
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
	details := NewServerDB()

	initializeUserRoutes(details)

	configureUserCollection(details.Db)
	godotenv.Load("../load.env")

	username := "Harry"
	user := entity.User{
		Username:     username,
		Fullname:     "Thompson " + username,
		Password:     "101" + username,
		Email:        username + "@email.com",
		MobileNumber: "0700",
	}

	userJson, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", "/user/create", bytes.NewBuffer(userJson))
	if err != nil {
		log.Fatalln("Could not create request")
	}

	recorder1 := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder1, req)
	require.Equal(t, 200, recorder1.Code)

	recorder2 := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder2, req)
	require.Equal(t, 400, recorder2.Code)

	require.NotEqual(t, recorder1.Body.Bytes(), recorder2.Body.Bytes())

	username = "Tom"
	user = entity.User{
		Username:     username,
		Fullname:     "Thompson " + username,
		Password:     "101" + username,
		Email:        username + "@email.com",
		MobileNumber: "0800",
	}

	userJson, _ = json.Marshal(user)
	req, err = http.NewRequest("POST", "/user/create", bytes.NewBuffer(userJson))
	if err != nil {
		log.Fatalln("Could not create request")
	}

	recorder1 = httptest.NewRecorder()
	details.Server.ServeHTTP(recorder1, req)
	require.Equal(t, 200, recorder1.Code)

	require.NoError(t, deleteRecords(details.Db, "users"))
	require.NoError(t, dropDatabase(details.Db))
}

func loginUserTest(t *testing.T, details *ServerDB, username string, triggers ...string) string {
	var lreq struct {
		Username string `json:"username" form:"username" binding:"required"`
		Password string `json:"password" form:"password" binding:"required"`
	}

	lreq.Username = username
	lreq.Password = "101" + username

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

	expectedBody := entity.UserResponse{}

	if len(triggers) > 1 {
		if triggers[1] == "unauthorised" {
			require.Equal(t, 401, recorder.Code)
			require.NotZero(t, recorder.Body.String)
		}
	} else {
		err = json.Unmarshal(recorder.Body.Bytes(), &expectedBody)
		require.NoError(t, err)
		require.Equal(t, 200, recorder.Code, expectedBody)
		require.Equal(t, "", expectedBody.MobileNumber)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.NotZero(t, expectedBody.ID)
		require.NotZero(t, expectedBody.Token)
		require.Equal(t, expectedBody.Username, expectedBody.Username)
		require.Equal(t, expectedBody.Fullname, expectedBody.Fullname)
		require.Equal(t, expectedBody.Email, expectedBody.Email)
		require.Equal(t, false, expectedBody.EmailIsVerfied)
		require.True(t, expectedBody.CreatedAt.Before(time.Now()))
	}
	return expectedBody.Token
}

func TestLoginUser(t *testing.T) {
	details := NewServerDB()

	initializeUserRoutes(details)

	username := "Harry"
	createUserTest(t, details, username)
	token := loginUserTest(t, details, username)

	require.NotZero(t, token)

	require.NoError(t, deleteRecords(details.Db, "users"))
	require.NoError(t, dropDatabase(details.Db))
}

func getUserTest(t *testing.T, details *ServerDB, user entity.UserResponse) entity.User {
	req, _ := http.NewRequest("GET", "/user/get", nil)
	req.Header.Add("Authorization", "Bearer "+user.Token)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)

	var gUser = entity.User{}
	err := json.Unmarshal(recorder.Body.Bytes(), &gUser)
	require.NoError(t, err)

	require.Equal(t, 200, recorder.Code)

	require.Equal(t, user.Username, gUser.Username)
	require.Equal(t, user.Fullname, gUser.Fullname)
	require.Equal(t, user.Email, gUser.Email)
	require.Equal(t, user.ID, gUser.ID)
	require.Equal(t, user.MobileNumber, gUser.MobileNumber)

	return gUser
}

func TestGetUser(t *testing.T) {
	details := NewServerDB()

	initializeUserRoutes(details)

	username := "Harry"
	user := createUserTest(t, details, username)

	_ = getUserTest(t, details, user)

	require.NoError(t, deleteRecords(details.Db, "users"))
	require.NoError(t, dropDatabase(details.Db))
}

func TestChangePassword(t *testing.T) {
	details := NewServerDB()

	initializeUserRoutes(details)

	username := "Harry"
	user := createUserTest(t, details, username)

	var cpReq struct {
		Password    string `json:"password" form:"password" binding:"required"`
		NewPassword string `json:"new_password" form:"new_password" binding:"required"`
	}

	cpReq.Password = "101" + username
	cpReq.NewPassword = username + "101"

	cpReqJson, _ := json.Marshal(cpReq)
	req, _ := http.NewRequest("PUT", "/user/password", bytes.NewBuffer(cpReqJson))
	req.Header.Add("Authorization", "Bearer "+user.Token)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)
	require.NotEqual(t, "404 page not found", recorder.Body.String())

	token := loginUserTest(t, details, username, "", "unauthorised")
	require.Empty(t, token)

	token = loginUserTest(t, details, username, "reversed")
	require.NotEmpty(t, token)

	require.NoError(t, deleteRecords(details.Db, "users"))
	require.NoError(t, dropDatabase(details.Db))
}

func updateUserTest(t *testing.T, details *ServerDB, token string, updateDetail string, user entity.User) {
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
	require.Equal(t, 200, recorder.Code)
	require.NotEqual(t, "404 page not found", recorder.Body.String())

	req, _ = http.NewRequest("GET", "/user/get", nil)
	req.Header.Add("Authorization", "Bearer "+token)

	recorder = httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)
	require.NotEqual(t, "404 page not found", recorder.Body.String())

	var updatedUser = entity.User{}
	err := json.Unmarshal(recorder.Body.Bytes(), &updatedUser)
	require.NoError(t, err)

	switch updateDetail {
	case "username":
		require.NotEqual(t, user.Username, updatedUser.Username)
	case "profile_picture":
		require.NotEqual(t, user.ProfilePicture, updatedUser.ProfilePicture)
	case "email":
		require.NotEqual(t, user.Email, updatedUser.Email)
	case "mobile_number":
		require.NotEqual(t, user.MobileNumber, updatedUser.MobileNumber)
	case "default_payment_method":
		require.NotEqual(t, user.DefaultPaymentMethod, updatedUser.DefaultPaymentMethod)
	}

	require.True(t, updatedUser.LastUpdated.After(user.LastUpdated))
}

func TestUpdateUserFlexible(t *testing.T) {
	details := NewServerDB()

	initializeUserRoutes(details)

	username := "Harry"
	user := createUserTest(t, details, username)

	gUser := getUserTest(t, details, user)

	// test updated profile picture
	updateUserTest(t, details, user.Token, "profile_picture", gUser)

	// test updated default payment method
	updateUserTest(t, details, user.Token, "default_payment_method", gUser)

	// test updated email
	updateUserTest(t, details, user.Token, "email", gUser)

	// test updated mobile number
	updateUserTest(t, details, user.Token, "mobile_number", gUser)

	// test updated username
	updateUserTest(t, details, user.Token, "username", gUser)

	require.NoError(t, deleteRecords(details.Db, "users"))
	require.NoError(t, dropDatabase(details.Db))
}

func addLocationTest(t *testing.T, details *ServerDB, user entity.UserResponse, location entity.Location, triggers ...string) {

	lreqJson, err := json.Marshal(location)
	require.NoError(t, err)

	req, _ := http.NewRequest("PUT", "/user/location/add", bytes.NewBuffer(lreqJson))
	req.Header.Add("Authorization", "Bearer "+user.Token)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)

	updatedUser := getUserTest(t, details, user)

	if len(triggers) > 0 {
		if triggers[0] == "second_location" {
			require.NotEqual(t, location, updatedUser.DefaultDeliveryLocation)
			require.NotEqual(t, location, updatedUser.RegisteredLocations[0])
			require.Equal(t, location, updatedUser.RegisteredLocations[1])
			require.Equal(t, location.HouseNumber, updatedUser.RegisteredLocations[1].HouseNumber)
			require.Equal(t, location.PhoneNo, updatedUser.RegisteredLocations[1].PhoneNo)
			require.Equal(t, location.Street, updatedUser.RegisteredLocations[1].Street)
			require.Equal(t, location.CityOrTown, updatedUser.RegisteredLocations[1].CityOrTown)
			require.Equal(t, location.State, updatedUser.RegisteredLocations[1].State)
			require.Equal(t, location.Country, updatedUser.RegisteredLocations[1].Country)
			require.Equal(t, location.ZipCode, updatedUser.RegisteredLocations[1].ZipCode)
		}
	} else {
		require.NotEmpty(t, updatedUser.RegisteredLocations)
		require.Equal(t, location, updatedUser.RegisteredLocations[0])
		require.Equal(t, location, updatedUser.DefaultDeliveryLocation)
		require.Equal(t, location.HouseNumber, updatedUser.RegisteredLocations[0].HouseNumber)
		require.Equal(t, location.PhoneNo, updatedUser.RegisteredLocations[0].PhoneNo)
		require.Equal(t, location.Street, updatedUser.RegisteredLocations[0].Street)
		require.Equal(t, location.CityOrTown, updatedUser.RegisteredLocations[0].CityOrTown)
		require.Equal(t, location.State, updatedUser.RegisteredLocations[0].State)
		require.Equal(t, location.Country, updatedUser.RegisteredLocations[0].Country)
		require.Equal(t, location.ZipCode, updatedUser.RegisteredLocations[0].ZipCode)
	}
}

func TestAddLocation(t *testing.T) {
	details := NewServerDB()

	initializeUserRoutes(details)

	username := "Harry"
	user := createUserTest(t, details, username)

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
	addLocationTest(t, details, user, location)

	location = entity.Location{
		HouseNumber: "50",
		PhoneNo:     "+2346446",
		Street:      "Mawuli",
		CityOrTown:  "Left City",
		State:       "Sroboth",
		Country:     "Land",
	}

	// Test for a second location
	addLocationTest(t, details, user, location, "second_location")

	require.NoError(t, deleteRecords(details.Db, "users"))
	require.NoError(t, dropDatabase(details.Db))
}
