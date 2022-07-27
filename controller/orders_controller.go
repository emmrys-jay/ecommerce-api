package controller

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	"github.com/Emmrys-Jay/ecommerce-api/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderProductRequest struct {
	Fullname      string          `json:"fullname" binding:"required"`
	ProductID     string          `json:"product_id" binding:"required"`
	Quantity      int             `json:"quantity" binding:"required"`
	Location      entity.Location `json:"location" binding:"required"`
	PaymentMethod string          `json:"payment_method" binding:"required"`
}

// OrderProduct serves an order-a-product request from a user directly
func (u *UserController) OrderProduct(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "orders")
	var req OrderProductRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	username, err := util.UsernameFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	_, productName, err := repository.OrderProductDirectly(
		collection,
		req.Location,
		req.Quantity,
		username,
		req.Fullname,
		req.ProductID,
		req.PaymentMethod,
	)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusBadRequest, req.ProductID)
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("you just bought %d product(s) with name - %s and id - %s", req.Quantity, productName, req.ProductID)
	ctx.JSON(http.StatusOK, gin.H{"success": response})
}

type GetOrdersWithUsernameResult struct {
	PageID        int            `json:"page_id"`
	ResultsFound  int            `json:"results_found"`
	NumberOfPages int            `json:"no_of_pages"`
	Data          []entity.Order `json:"data"`
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

	username, err := util.UsernameFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	orders, length, err := repository.GetOrdersWithUsername(collection, username, param.Limit, param.Offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := GetOrdersWithUsernameResult{
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

func (u *UserController) DeliverOrder(ctx *gin.Context) {
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

	_, err = repository.DeliverOrder(collection, orderID, username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"response": "success!"})
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
			ctx.JSON(http.StatusNotFound, util.ErrorResponse(err))
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

	username, err := util.UsernameFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	numProductsOrdered, namesProductsOrdered, err := repository.OrderAllCartItems(
		collection,
		username,
		req.Fullname,
		req.PaymentMethod,
		req.Location,
	)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("successfully ordered %d products with names: %s", numProductsOrdered, namesProductsOrdered)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}
