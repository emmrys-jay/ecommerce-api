package controller

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	"github.com/Emmrys-Jay/ecommerce-api/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderProductRequest struct {
	Fullname      string          `json:"fullname" binding:"required"`
	Quantity      int             `json:"quantity" binding:"required"`
	Location      entity.Location `json:"location" binding:"required"`
	PaymentMethod string          `json:"payment_method" binding:"required"`
}

type OrderProductResult struct {
	Response string `json:"response"`
	OrderID  string `json:"order-id"`
}

// OrderProduct serves an order-a-product request from a user directly
func (u *UserController) OrderProduct(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "orders")
	var req OrderProductRequest

	productID := ctx.Param("productID")
	if productID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid param - product not specified"})
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	userID, err := util.UserIDFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	orderID := primitive.NewObjectIDFromTimestamp(time.Now()).Hex()
	_, productName, err := repository.OrderProductDirectly(
		collection,
		&req.Location,
		req.Quantity,
		userID,
		req.Fullname,
		productID,
		req.PaymentMethod,
	)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusBadRequest, productID)
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("you just bought %d product(s) with name - %s and id - %s", req.Quantity, productName, productID)
	fResponse := OrderProductResult{
		Response: response,
		OrderID:  orderID,
	}

	ctx.JSON(http.StatusOK, fResponse)
}

// GetOrder returns an order db entry
func (u *UserController) GetOrder(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "orders")

	orderID := ctx.Param("order-ID")
	if orderID == "" {
		ctx.JSON(http.StatusBadRequest, "Invalid Param - order ID not specified")
		return
	}

	order, err := repository.GetSingleOrder(collection, orderID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "order specified does not exist"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, order)
}

// GetOrdersWithUsername gets orders associated with a username
func (u *UserController) GetOrdersWithUsername(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "orders")
	pageSize, pageID := 5, 1
	var err error

	pageIDString := ctx.Query("page_id")

	if pageIDString != "" {
		pageID, err = strconv.Atoi(pageIDString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
			return
		}

		if pageID < 1 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid url param"})
			return
		}
	}

	var param = struct {
		Offset int
		Limit  int
	}{
		Offset: pageSize * (pageID - 1),
		Limit:  pageSize,
	}

	userID, err := util.UserIDFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	orders, length, err := repository.GetOrdersByUser(collection, userID, param.Limit, param.Offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := entity.PaginationResponse{
		PageID:        pageID,
		NumberOfPages: int(math.Ceil(float64(length) / float64(pageSize))),
		ResultsFound:  int(length),
		Data:          orders,
	}

	if response.NumberOfPages < 1 {
		response.PageID = 0
	}

	ctx.JSON(http.StatusOK, response)
}

func (u *UserController) ReceiveOrder(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "orders")

	orderID := ctx.Param("order-id")
	if orderID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid url param"})
		return
	}

	username, err := util.UsernameFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	_, err = repository.ReceiveOrder(collection, username, orderID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"response": "success!"})
}

type OrderAllCartItemsRequest struct {
	Fullname      string          `json:"fullname" binding:"required"`
	Location      entity.Location `json:"location" binding:"required"`
	PaymentMethod string          `json:"payment_method" binding:"required"`
}

// OrderAllCartItems orders all items currently stored in a users cart
func (u *UserController) OrderAllCartItems(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "orders")
	var req OrderAllCartItemsRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	userID, err := util.UserIDFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	numProductsOrdered, err := repository.OrderAllCartItems(
		collection,
		userID,
		req.Fullname,
		req.PaymentMethod,
		req.Location,
	)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("successfully ordered %d products", numProductsOrdered)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}
