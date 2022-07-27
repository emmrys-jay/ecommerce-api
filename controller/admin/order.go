package controller

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	util "github.com/Emmrys-Jay/ecommerce-api/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type GetAllOrdersResult struct {
	PageID        int            `json:"page_id"`
	ResultsFound  int            `json:"results_found"`
	NumberOfPages int            `json:"no_of_pages"`
	Data          []entity.Order `json:"data"`
}

// GetAllOrders handles a request to get all site orders from an admin
func (a *AdminController) GetAllOrders(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "orders")
	var err error
	var pageID int
	var pageSize = 5

	pageIDString := ctx.Query("page_id")
	if pageIDString == "" {
		pageID = 1
	} else {
		pageID, err = strconv.Atoi(pageIDString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"response": "invalid params - page_id"})
			return
		}
	}

	if pageID <= 0 {
		pageID = 1
	}

	var params = struct {
		Limit  int
		Offset int
	}{
		Limit:  pageSize,
		Offset: pageSize * (pageID - 1),
	}

	orders, length, err := repository.GetAllOrders(collection, params.Limit, params.Offset)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := GetAllOrdersResult{
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

// DeleteCartItem is an admin specific handler to delete a single item in cart
func (a *AdminController) DeleteOrder(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "orders")

	orderID := ctx.Param("id")
	if orderID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid param - no id specified"})
		return
	}

	_, err := repository.DeleteOrder(collection, orderID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("deleted order with id - %s", orderID)
	ctx.JSON(http.StatusOK, gin.H{"success": response})
}

// DeleteAllCartItems is an admin specific handler to delete all items in cart of different users
func (a *AdminController) DeleteAllOrdersWithUsername(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "orders")

	username := ctx.Param("username")
	if username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid param - no id specified"})
		return
	}

	result, err := repository.DeleteAllOrdersWithUsername(collection, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("deleted %d orders", result.DeletedCount)
	ctx.JSON(http.StatusOK, gin.H{"success": response})
}

// DeleteAllCartItems is an admin specific handler to delete all items in cart of different users
func (a *AdminController) DeleteAllOrders(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "orders")

	result, err := repository.DeleteAllOrders(collection)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("deleted %d orders", result.DeletedCount)
	ctx.JSON(http.StatusOK, gin.H{"success": response})
}
