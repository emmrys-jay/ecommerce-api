package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func initializeCartRoutes(details *ServerDB) {
	usercontroller := NewUserController(details.Db)
	cart := details.Server.Group("/user/cart")
	{
		cart.POST("/add", usercontroller.AddToCart)
		cart.DELETE("/remove/:cart-id", usercontroller.RemoveFromCart)
		cart.PUT("/update", usercontroller.UpdateCartQuantity)
		cart.GET("/getall", usercontroller.GetUserCartItems)
	}
}

func getCartItem(t *testing.T, details *ServerDB, quantity int64, productName, username string, trigger ...string) (*entity.CartItem, error) {
	collection := details.Db.Collection("cart")

	var cart entity.CartItem

	filter := bson.M{"product_name": productName}
	err := collection.FindOne(context.Background(), filter).Decode(&cart)
	if len(trigger) > 0 {
		if trigger[0] == "deleted" {
			assert.Error(t, err)
			assert.Equal(t, mongo.ErrNoDocuments, err)

			return nil, err
		} else {
			return nil, fmt.Errorf("Invalid trigger input")
		}
	} else {
		assert.NoError(t, err)
		assert.Equal(t, productName, cart.ProductName)
		assert.Equal(t, username, cart.Username)
		assert.Equal(t, quantity, cart.Quantity)
		assert.NotZero(t, cart.ID)
		assert.NotZero(t, cart.Product)
		assert.True(t, cart.DateAdded.Before(time.Now()))
	}

	return &cart, nil
}

func addProductToCart(t *testing.T, details *ServerDB, user UserResponse, productID string, quantity int64) {
	aReq := AddToCartRequest{
		ProductID: productID,
		Quantity:  int64(quantity),
	}
	aReqJson, _ := json.Marshal(aReq)
	req, err := http.NewRequest("POST", "/user/cart/add", bytes.NewBuffer(aReqJson))
	req.Header.Add("Authorization", "Bearer "+user.Token)
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code)
	assert.NotEqual(t, "404 page not found", recorder.Body.String())
}

func TestAddToCart(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}
	initializeCartRoutes(&details)
	initializeUserRoutes(&details)

	username := "Harry"
	user := createUserTest(t, &details, username)

	product := createProduct(t, &details, "Chandler Bags")

	var quantity int64 = 90
	addProductToCart(t, &details, user, product.ID, quantity)

	_, err := getCartItem(t, &details, quantity, product.Name, user.Username)
	assert.NoError(t, err)

	deleteRecords(details.Db, "cart")
	dropDatabase(details.Db)
}

func TestRemoveFromCart(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}
	initializeCartRoutes(&details)
	initializeUserRoutes(&details)

	username := "Harry"
	user := createUserTest(t, &details, username)
	product := createProduct(t, &details, "Chandler Bags")

	var quantity int64 = 90
	addProductToCart(t, &details, user, product.ID, quantity)
	cartItem, err := getCartItem(t, &details, quantity, product.Name, user.Username)
	assert.NotNil(t, cartItem)
	assert.NoError(t, err)

	path := fmt.Sprintf("/user/cart/remove/%s", cartItem.ID)
	req, err := http.NewRequest("DELETE", path, nil)
	req.Header.Add("Authorization", "Bearer "+user.Token)
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code)
	assert.NotEqual(t, "404 page not found", recorder.Body.String())

	cartItem, err = getCartItem(t, &details, quantity, product.Name, user.Username, "deleted")
	assert.Nil(t, cartItem)
	assert.Error(t, err)

	deleteRecords(details.Db, "cart")
	dropDatabase(details.Db)
}

func TestUpdateCartQuantity(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}
	initializeCartRoutes(&details)
	initializeUserRoutes(&details)

	username := "Harry"
	user := createUserTest(t, &details, username)
	product := createProduct(t, &details, "Chandler Bags")

	var quantity int64 = 90
	addProductToCart(t, &details, user, product.ID, quantity)

	cartItem, err := getCartItem(t, &details, quantity, product.Name, user.Username)
	assert.NotNil(t, cartItem)
	assert.NoError(t, err)

	uReq := UpdateCartRequest{
		CartID:   cartItem.ID,
		Quantity: int(quantity * 3),
	}

	uReqJson, _ := json.Marshal(uReq)
	req, err := http.NewRequest("PUT", "/user/cart/update", bytes.NewBuffer(uReqJson))
	req.Header.Add("Authorization", "Bearer "+user.Token)
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code)
	assert.NotEqual(t, "404 page not found", recorder.Body.String())

	deleteRecords(details.Db, "cart")
	dropDatabase(details.Db)
}

func TestGetAllUserCartItems(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}
	initializeCartRoutes(&details)
	initializeUserRoutes(&details)

	productIDs := []string{}
	for _, v := range productNames {
		product := createProduct(t, &details, v)
		productIDs = append(productIDs, product.ID)
	}

	username := "Harry"
	user := createUserTest(t, &details, username)
	assert.NotZero(t, user)

	for _, val := range productIDs {
		addProductToCart(t, &details, user, val, rand.Int63n(100))
	}
	req, err := http.NewRequest("GET", "/user/cart/getall", nil)
	req.Header.Add("Authorization", "Bearer "+user.Token)
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code)
	assert.NotEqual(t, "404 page not found", recorder.Body.String())

	var result GetUserCartItemsResult
	err = json.Unmarshal(recorder.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, len(productIDs), result.ResultsFound)
	assert.Equal(t, len(productNames), len(result.Data))
	assert.Equal(t, 1, result.PageID)
	assert.NotZero(t, result.NumberOfPages)
	for _, v := range result.Data {
		assert.NotZero(t, v.ID)
		assert.NotZero(t, v.Username)
		assert.NotZero(t, v.Product)
		assert.NotZero(t, v.ProductName)
		assert.NotZero(t, v.Quantity)
		assert.True(t, v.DateAdded.Before(time.Now()))
	}

	deleteRecords(details.Db, "cart")
	dropDatabase(details.Db)
}
