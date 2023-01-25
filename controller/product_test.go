// product_test

package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/util"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var productNames = [...]string{
	"Charis Bags",
	"Chandlers Rags",
	"Chris Rugs",
	"chuckles Beads",
	"Laundry Chairs",
}

func createProduct(t *testing.T, details *ServerDB, productName string, triggers ...string) *entity.Product {

	product := entity.Product{
		ID:          primitive.NewObjectID().String()[10:34],
		Name:        productName,
		Price:       386520.0,
		Currency:    "CAD",
		Quantity:    216348,
		Description: util.RandomString(),
		Category:    util.RandomString(),
	}

	if len(triggers) > 0 {
		product.Category = triggers[0]

	}

	result, err := details.Db.Collection("products").InsertOne(context.Background(), product)
	require.NoError(t, err)
	require.NotZero(t, result.InsertedID)

	return &product
}

func TestFindProducts(t *testing.T) {
	details := NewServerDB()

	initializeProductRoutes(details)

	for _, v := range productNames {
		createProduct(t, details, v)
	}

	fpReq := FindProductsRequest{
		Name:     "Ch",
		PageID:   1,
		PageSize: 1, //Too small number, should be changed to 5
	}

	fpReqJson, _ := json.Marshal(fpReq)
	req, err := http.NewRequest("GET", "/products/find", bytes.NewBuffer(fpReqJson))
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)

	var result = FindProductsResult{}
	json.Unmarshal(recorder.Body.Bytes(), &result)

	require.Equal(t, 5, len(result.Data))
	require.Equal(t, int64(1), result.PageID)
	require.Equal(t, int64(1), result.NumberOfPages)
	require.Equal(t, int64(5), result.ResultsFound)

	for _, v := range result.Data {
		require.NotZero(t, v.Name)
		require.NotZero(t, v.Price)
		require.NotZero(t, v.Currency)
		require.NotZero(t, v.Quantity)
		require.NotZero(t, v.Description)
		require.NotZero(t, v.Category)
		require.True(t, v.LastUpdated.Before(time.Now()))
	}

	deleteRecords(details.Db, "products")
	dropDatabase(details.Db)
}

func TestFindOneProduct(t *testing.T) {
	details := NewServerDB()

	initializeProductRoutes(details)

	productName := "Chandlers Rags"
	product := createProduct(t, details, productName)

	path := fmt.Sprintf("/products/findone/%s", product.ID)
	req, err := http.NewRequest("GET", path, nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)

	var result = entity.Product{}
	err = json.Unmarshal(recorder.Body.Bytes(), &result)
	require.NoError(t, err)

	require.Equal(t, product.Name, result.Name)
	require.Equal(t, product.ID, result.ID)
	require.Equal(t, product.Price, result.Price)
	require.Equal(t, product.Quantity, result.Quantity)
	require.Equal(t, product.Category, result.Category)
	require.Equal(t, product.Description, result.Description)

	deleteRecords(details.Db, "products")
	dropDatabase(details.Db)
}

func addReviewTest(t *testing.T, details *ServerDB, stars int, triggers ...string) {
	productName := "Chandlers Rags"
	product := createProduct(t, details, productName)

	username := "Harry"
	user := createUserTest(t, details, username)

	review := entity.Review{
		User:      username,
		Stars:     int64(stars),
		Comment:   "Great product",
		CreatedAt: time.Now(),
	}

	if len(triggers) > 0 {
		if triggers[0] == "stars" {
			review.Stars = int64(stars)
		}
	}

	rJson, _ := json.Marshal(review)
	path := fmt.Sprintf("/products/%s/addreview", product.ID)
	req, err := http.NewRequest("PUT", path, bytes.NewBuffer(rJson))
	req.Header.Add("Authorization", "Bearer "+user.Token)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)

	if len(triggers) > 0 {
		if triggers[0] == "stars" {
			require.Equal(t, 400, recorder.Code)
			require.NotEqual(t, "404 page not found", recorder.Body.String())
		}
	} else {
		require.Equal(t, 200, recorder.Code)
		require.NotEqual(t, "404 page not found", recorder.Body.String())
	}

}

func TestAddReview(t *testing.T) {
	details := NewServerDB()

	initializeProductRoutes(details)
	initializeUserRoutes(details)

	// Test with a valid number of stars (1 - 5)
	addReviewTest(t, details, 5)

	// Test with an invalid number of stars (>5)
	addReviewTest(t, details, 9, "stars")

	// Test with an invalid number of stars (<1)
	addReviewTest(t, details, -3, "stars")

	deleteRecords(details.Db, "products")
	dropDatabase(details.Db)
}

func TestGetProductsByCategory(t *testing.T) {
	details := NewServerDB()

	initializeProductRoutes(details)

	productsCategory := "bags"
	for _, v := range productNames {
		createProduct(t, details, v, productsCategory)
	}

	path := fmt.Sprintf("/products/get/%s", productsCategory)
	req, err := http.NewRequest("GET", path, nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)
	require.NotEqual(t, "404 page not found", recorder.Body.String())

	var result FindProductsResult
	err = json.Unmarshal(recorder.Body.Bytes(), &result)
	require.NoError(t, err)

	require.Equal(t, 5, len(result.Data))
	require.Equal(t, int64(1), result.PageID)
	require.Equal(t, int64(1), result.NumberOfPages)
	require.Equal(t, int64(5), result.ResultsFound)

	for _, v := range result.Data {
		require.NotZero(t, v.Name)
		require.NotZero(t, v.Price)
		require.NotZero(t, v.Currency)
		require.NotZero(t, v.Quantity)
		require.NotZero(t, v.Description)
		require.Equal(t, productsCategory, v.Category)
		require.True(t, v.LastUpdated.Before(time.Now()))
	}

	deleteRecords(details.Db, "products")
	dropDatabase(details.Db)
}
