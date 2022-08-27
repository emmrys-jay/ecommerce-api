package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func initializeOrdersRoutes(details *ServerDB) {
	userController := NewUserController(details.Db)

	orders := details.Server.Group("/products/order")
	{
		orders.POST("/:productID", userController.OrderProduct)
		orders.GET("/get/:order-ID", userController.GetOrder)
		orders.GET("/get", userController.GetOrdersWithUsername)
		orders.PUT("/receive/:order-id", userController.ReceiveOrder)
		orders.POST("/cart", userController.OrderAllCartItems)
	}
}

func orderProductTest(t *testing.T, details *ServerDB, user UserResponse, productID string) string {
	oReq := OrderProductRequest{
		Fullname:      user.Username,
		Quantity:      9,
		PaymentMethod: "nil",
		Location: entity.Location{
			HouseNumber: "77",
			CityOrTown:  "My Town",
			Street:      "Ajao",
			State:       "Lagos",
			Country:     "Nigeria",
		},
	}

	oReqJson, _ := json.Marshal(oReq)

	path := fmt.Sprintf("/products/order/%s", productID)
	req, err := http.NewRequest("POST", path, bytes.NewBuffer(oReqJson))
	req.Header.Add("Authorization", "Bearer "+user.Token)
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code)

	var result OrderProductResult
	err = json.Unmarshal(recorder.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.NotZero(t, result.Response)
	assert.NotZero(t, result.OrderID)

	return result.OrderID
}

func getOrderTest(t *testing.T, details *ServerDB, user UserResponse, expectedOutputLength int, triggers ...string) (*entity.Order, error) {
	var req *http.Request
	var err error

	if len(triggers) < 2 {
		return nil, fmt.Errorf("Trigger not specified")
	} else {
		if triggers[0] == "id" {
			path := fmt.Sprintf("/products/order/get/%s", triggers[1])
			req, err = http.NewRequest("GET", path, nil)
			assert.NoError(t, err)
		} else if triggers[0] == "username" {
			req, err = http.NewRequest("GET", "/products/order/get", nil)
			req.Header.Add("Authorization", "Bearer "+user.Token)
			assert.NoError(t, err)
		} else {
			return nil, fmt.Errorf("Invalid trigger specified")
		}
	}

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code)

	if expectedOutputLength == 1 {
		var order entity.Order
		err = json.Unmarshal(recorder.Body.Bytes(), &order)
		assert.NoError(t, err)

		assert.NotZero(t, order.ID)
		assert.NotZero(t, order.Username)
		assert.NotZero(t, order.FullName)
		assert.NotZero(t, order.DeliveryLocation)
		assert.NotZero(t, order.ProductQuantity)
		assert.NotZero(t, order.Product.Price)
		assert.True(t, order.CreatedAt.Before(time.Now()))

		return &order, nil
	} else if expectedOutputLength > 1 {
		var result GetOrdersWithUsernameResult
		err = json.Unmarshal(recorder.Body.Bytes(), &result)
		assert.NoError(t, err)

		assert.Equal(t, expectedOutputLength, result.ResultsFound)
		assert.Equal(t, expectedOutputLength, len(result.Data))
		assert.Equal(t, 1, result.PageID)

		for _, v := range result.Data {
			assert.NotZero(t, v.ID)
			assert.NotZero(t, v.Username)
			assert.NotZero(t, v.FullName)
			assert.NotZero(t, v.DeliveryLocation)
			assert.NotZero(t, v.ProductQuantity)
			assert.NotZero(t, v.Product.Price)
			assert.True(t, v.CreatedAt.Before(time.Now()))
		}
	} else {
		return nil, fmt.Errorf("Invalid Input into expected output length")
	}
	return nil, nil
}

func TestOrderProduct(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}
	initializeOrdersRoutes(&details)
	initializeUserRoutes(&details)

	product := createProduct(t, &details, "Chandlers Bags")

	username := "Harry"
	user := createUserTest(t, &details, username)
	assert.NotZero(t, user)

	orderID := orderProductTest(t, &details, user, product.ID)
	assert.NotZero(t, orderID)

	_, err := getOrderTest(t, &details, user, 1, "id", orderID)
	assert.NoError(t, err)

	deleteRecords(details.Db, "orders")
	dropDatabase(details.Db)

}

func TestGetOrderWithUsername(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}
	initializeOrdersRoutes(&details)
	initializeUserRoutes(&details)

	productIDs := []string{}
	for _, v := range productNames {
		product := createProduct(t, &details, v)
		productIDs = append(productIDs, product.ID)
	}

	username := "Harry"
	user := createUserTest(t, &details, username)
	assert.NotZero(t, user)

	for _, v := range productIDs {
		orderID := orderProductTest(t, &details, user, v)
		assert.NotZero(t, orderID)
	}

	_, err := getOrderTest(t, &details, user, 5, "username", username)
	assert.NoError(t, err)

	deleteRecords(details.Db, "orders")
	dropDatabase(details.Db)
}

//func TestDeliverOrder(t *testing.T) {
//	details := ServerDB{
//		Db:     connectDB(),
//		Server: gin.Default(),
//	}
//	initializeOrdersRoutes(&details)
//	initializeUserRoutes(&details)
//
//	product := createProduct(t, &details, "Chandlers Bags")
//
//	username := "Harry"
//	user := createUserTest(t, &details, username)
//	assert.NotZero(t, user)
//
//	orderID := orderProductTest(t, &details, user, product.ID)
//	assert.NotZero(t, orderID)
//
//	path := fmt.Sprintf("/products/order/deliver/%s", orderID)
//	req, err := http.NewRequest("PUT", path, nil)
//	req.Header.Add("Authorization", "Bearer "+user.Token)
//	assert.NoError(t, err)
//
//	recorder := httptest.NewRecorder()
//	details.Server.ServeHTTP(recorder, req)
//	assert.Equal(t, 200, recorder.Code)
//
//	order, err := getOrderTest(t, &details, user, 1, "id", orderID)
//	assert.NoError(t, err)
//	assert.True(t, order.IsDelivered)
//
//	deleteRecords(details.Db, "orders")
//	dropDatabase(details.Db)
//}

func TestReceiveOrder(t *testing.T) {
	details := ServerDB{
		Db:     connectDB(),
		Server: gin.Default(),
	}
	initializeOrdersRoutes(&details)
	initializeUserRoutes(&details)

	product := createProduct(t, &details, "Chandlers Bags")

	username := "Harry"
	user := createUserTest(t, &details, username)
	assert.NotZero(t, user)

	orderID := orderProductTest(t, &details, user, product.ID)
	assert.NotZero(t, orderID)

	path := fmt.Sprintf("/products/order/receive/%s", orderID)
	req, err := http.NewRequest("PUT", path, nil)
	req.Header.Add("Authorization", "Bearer "+user.Token)
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code)

	order, err := getOrderTest(t, &details, user, 1, "id", orderID)
	assert.NoError(t, err)
	assert.True(t, order.IsReceived)

	deleteRecords(details.Db, "orders")
	dropDatabase(details.Db)
}
