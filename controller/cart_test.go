package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func getCartItem(t *testing.T, details *ServerDB, quantity int64, productName, username string, trigger ...string) (*entity.CartItem, error) {
	collection := details.Db.Collection("cart")

	var cart entity.CartItem

	filter := bson.M{"product_name": productName}
	err := collection.FindOne(context.Background(), filter).Decode(&cart)
	if len(trigger) > 0 {
		if trigger[0] == "deleted" {
			require.Error(t, err)
			require.Equal(t, mongo.ErrNoDocuments, err)

			return nil, err
		} else {
			return nil, fmt.Errorf("Invalid trigger input")
		}
	} else {
		require.NoError(t, err)
		require.Equal(t, productName, cart.ProductName)
		require.Equal(t, username, cart.Username)
		require.Equal(t, quantity, cart.Quantity)
		require.NotZero(t, cart.ID)
		require.NotZero(t, cart.Product)
		require.True(t, cart.DateAdded.Before(time.Now()))
	}

	return &cart, nil
}

func addProductToCart(t *testing.T, details *ServerDB, user UserResponse, productID string, quantity int64, trigger ...string) {
	aReq := AddToCartRequest{
		ProductID: productID,
		Quantity:  int64(quantity),
	}
	aReqJson, _ := json.Marshal(aReq)
	req, err := http.NewRequest("POST", "/user/cart/add", bytes.NewBuffer(aReqJson))
	req.Header.Add("Authorization", "Bearer "+user.Token)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)

	if len(trigger) > 0 {
		if trigger[0] == "unique" {
			require.Equal(t, 400, recorder.Code)
			require.NotEqual(t, "404 page not found", recorder.Body.String())
			return
		}
	}
	require.Equal(t, 200, recorder.Code)
	require.NotEqual(t, "404 page not found", recorder.Body.String())
}

func TestAddToCart(t *testing.T) {
	details := NewServerDB()

	initializeCartRoutes(details)
	initializeUserRoutes(details)

	username := "Harry"
	user := createUserTest(t, details, username)

	product := createProduct(t, details, "Chandler Bags")

	var quantity int64 = 90
	addProductToCart(t, details, user, product.ID, quantity)

	_, err := getCartItem(t, details, quantity, product.Name, user.Username)
	require.NoError(t, err)

	// Test Unique product name field in cart
	addProductToCart(t, details, user, product.ID, quantity, "unique")

	deleteRecords(details.Db, "cart")
	dropDatabase(details.Db)
}

func TestRemoveFromCart(t *testing.T) {
	details := NewServerDB()

	initializeCartRoutes(details)
	initializeUserRoutes(details)

	username := "Harry"
	user := createUserTest(t, details, username)
	product := createProduct(t, details, "Chandler Bags")

	var quantity int64 = 90
	addProductToCart(t, details, user, product.ID, quantity)
	cartItem, err := getCartItem(t, details, quantity, product.Name, user.Username)
	require.NotNil(t, cartItem)
	require.NoError(t, err)

	path := fmt.Sprintf("/user/cart/remove/%s", cartItem.ID)
	req, err := http.NewRequest("DELETE", path, nil)
	req.Header.Add("Authorization", "Bearer "+user.Token)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)
	require.NotEqual(t, "404 page not found", recorder.Body.String())

	cartItem, err = getCartItem(t, details, quantity, product.Name, user.Username, "deleted")
	require.Nil(t, cartItem)
	require.Error(t, err)

	deleteRecords(details.Db, "cart")
	dropDatabase(details.Db)
}

func TestUpdateCartQuantity(t *testing.T) {
	details := NewServerDB()

	initializeCartRoutes(details)
	initializeUserRoutes(details)

	username := "Harry"
	user := createUserTest(t, details, username)
	product := createProduct(t, details, "Chandler Bags")

	var quantity int64 = 90
	addProductToCart(t, details, user, product.ID, quantity)

	cartItem, err := getCartItem(t, details, quantity, product.Name, user.Username)
	require.NotNil(t, cartItem)
	require.NoError(t, err)

	uReq := UpdateCartRequest{
		CartID:   cartItem.ID,
		Quantity: int(quantity * 3),
	}

	uReqJson, _ := json.Marshal(uReq)
	req, err := http.NewRequest("PUT", "/user/cart/update", bytes.NewBuffer(uReqJson))
	req.Header.Add("Authorization", "Bearer "+user.Token)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)
	require.NotEqual(t, "404 page not found", recorder.Body.String())

	deleteRecords(details.Db, "cart")
	dropDatabase(details.Db)
}

func TestGetAllUserCartItems(t *testing.T) {
	details := NewServerDB()

	initializeCartRoutes(details)
	initializeUserRoutes(details)

	productIDs := []string{}
	for _, v := range productNames {
		product := createProduct(t, details, v)
		productIDs = append(productIDs, product.ID)
	}

	username := "Harry"
	user := createUserTest(t, details, username)
	require.NotZero(t, user)

	for _, val := range productIDs {
		addProductToCart(t, details, user, val, rand.Int63n(100))
	}
	req, err := http.NewRequest("GET", "/user/cart/getall", nil)
	req.Header.Add("Authorization", "Bearer "+user.Token)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)
	require.NotEqual(t, "404 page not found", recorder.Body.String())

	var result GetUserCartItemsResult
	err = json.Unmarshal(recorder.Body.Bytes(), &result)
	require.NoError(t, err)

	require.Equal(t, len(productIDs), result.ResultsFound)
	require.Equal(t, len(productNames), len(result.Data))
	require.Equal(t, 1, result.PageID)
	require.NotZero(t, result.NumberOfPages)
	for _, v := range result.Data {
		require.NotZero(t, v.ID)
		require.NotZero(t, v.Username)
		require.NotZero(t, v.Product)
		require.NotZero(t, v.ProductName)
		require.NotZero(t, v.Quantity)
		require.True(t, v.DateAdded.Before(time.Now()))
	}

	deleteRecords(details.Db, "cart")
	dropDatabase(details.Db)
}
