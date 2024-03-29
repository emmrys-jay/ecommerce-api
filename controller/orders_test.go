// orders_test

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
	"github.com/stretchr/testify/require"
)

func orderProductTest(t *testing.T, details *ServerDB, user entity.UserResponse, productID string) string {
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
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)

	var result OrderProductResult
	err = json.Unmarshal(recorder.Body.Bytes(), &result)
	require.NoError(t, err)
	require.NotZero(t, result.Response)
	require.NotZero(t, result.OrderID)

	return result.OrderID
}

func getOrderTest(t *testing.T, details *ServerDB, user entity.UserResponse, expectedOutputLength int, triggers ...string) (*entity.Order, error) {
	var req *http.Request
	var err error

	if len(triggers) < 2 {
		return nil, fmt.Errorf("Trigger not specified")
	} else {
		if triggers[0] == "id" {
			path := fmt.Sprintf("/products/order/get/%s", triggers[1])
			req, err = http.NewRequest("GET", path, nil)
			require.NoError(t, err)
		} else if triggers[0] == "username" {
			req, err = http.NewRequest("GET", "/products/order/get", nil)
			req.Header.Add("Authorization", "Bearer "+user.Token)
			require.NoError(t, err)
		} else {
			return nil, fmt.Errorf("Invalid trigger specified")
		}
	}

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)

	if expectedOutputLength == 1 {
		var order entity.Order
		err = json.Unmarshal(recorder.Body.Bytes(), &order)
		require.NoError(t, err)

		require.NotZero(t, order.ID)
		require.NotZero(t, order.UserID)
		require.NotZero(t, order.FullName)
		require.NotZero(t, order.DeliveryLocation)
		require.NotZero(t, order.ProductQuantity)
		require.NotZero(t, order.Product.Price)
		require.True(t, order.CreatedAt.Before(time.Now()))

		return &order, nil
	} else if expectedOutputLength > 1 {
		var result entity.PaginationResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &result)
		require.NoError(t, err)

		require.Equal(t, expectedOutputLength, result.ResultsFound)
		require.NotNil(t, result.Data)
		require.Equal(t, 1, result.PageID)

	} else {
		return nil, fmt.Errorf("Invalid Input into expected output length")
	}
	return nil, nil
}

func TestOrderProduct(t *testing.T) {
	details := NewServerDB()

	initializeOrdersRoutes(details)
	initializeUserRoutes(details)

	product := createProduct(t, details, "Chandlers Bags")

	username := "Harry"
	user := createUserTest(t, details, username)
	require.NotZero(t, user)

	orderID := orderProductTest(t, details, user, product.ID)
	require.NotZero(t, orderID)

	_, err := getOrderTest(t, details, user, 1, "id", orderID)
	require.NoError(t, err)

	deleteRecords(details.Db, "orders")
	dropDatabase(details.Db)

}

func TestGetOrderWithUsername(t *testing.T) {
	details := NewServerDB()

	initializeOrdersRoutes(details)
	initializeUserRoutes(details)

	productIDs := []string{}
	for _, v := range productNames {
		product := createProduct(t, details, v)
		productIDs = append(productIDs, product.ID)
	}

	username := "Harry"
	user := createUserTest(t, details, username)
	require.NotZero(t, user)

	for _, v := range productIDs {
		orderID := orderProductTest(t, details, user, v)
		require.NotZero(t, orderID)
	}

	_, err := getOrderTest(t, details, user, 5, "username", username)
	require.NoError(t, err)

	deleteRecords(details.Db, "orders")
	dropDatabase(details.Db)
}

//func TestDeliverOrder(t *testing.T) {
//	details := ServerDB{
//		Db:     connectDB(),
//		Server: gin.Default(),
//	}
//	initializeOrdersRoutes(details)
//	initializeUserRoutes(details)
//
//	product := createProduct(t, details, "Chandlers Bags")
//
//	username := "Harry"
//	user := createUserTest(t, details, username)
//	require.NotZero(t, user)
//
//	orderID := orderProductTest(t, details, user, product.ID)
//	require.NotZero(t, orderID)
//
//	path := fmt.Sprintf("/products/order/deliver/%s", orderID)
//	req, err := http.NewRequest("PUT", path, nil)
//	req.Header.Add("Authorization", "Bearer "+user.Token)
//	require.NoError(t, err)
//
//	recorder := httptest.NewRecorder()
//	details.Server.ServeHTTP(recorder, req)
//	require.Equal(t, 200, recorder.Code)
//
//	order, err := getOrderTest(t, details, user, 1, "id", orderID)
//	require.NoError(t, err)
//	require.True(t, order.IsDelivered)
//
//	deleteRecords(details.Db, "orders")
//	dropDatabase(details.Db)
//}

func TestReceiveOrder(t *testing.T) {
	details := NewServerDB()

	initializeOrdersRoutes(details)
	initializeUserRoutes(details)

	product := createProduct(t, details, "Chandlers Bags")

	username := "Harry"
	user := createUserTest(t, details, username)
	require.NotZero(t, user)

	orderID := orderProductTest(t, details, user, product.ID)
	require.NotZero(t, orderID)

	path := fmt.Sprintf("/products/order/receive/%s", orderID)
	req, err := http.NewRequest("PUT", path, nil)
	req.Header.Add("Authorization", "Bearer "+user.Token)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	details.Server.ServeHTTP(recorder, req)
	require.Equal(t, 200, recorder.Code)

	order, err := getOrderTest(t, details, user, 1, "id", orderID)
	require.NoError(t, err)
	require.True(t, order.IsReceived)

	deleteRecords(details.Db, "orders")
	dropDatabase(details.Db)
}
