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

// GetCartItem gets a single product stored in a users cart, mainly used by admin
func (a *AdminController) GetCartItem(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "cart")

	cartID := ctx.Param("cart-id")
	if cartID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params"})
		return
	}

	cartItem, err := repository.GetCartItem(collection, cartID, "")
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusInternalServerError, gin.H{"not found": "No product in your cart matches the params specified"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, cartItem)
}

type GetAllUserCartResult struct {
	PageID        int               `json:"page_id"`
	ResultsFound  int               `json:"results_found"`
	NumberOfPages int               `json:"no_of_pages"`
	Data          []entity.CartItem `json:"data"`
}

// GetAllCartItems gets a single product stored in a users cart, mainly used by admin
func (a *AdminController) GetAllCartItems(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "cart")
	var pageID, pageSize = 1, 5
	var err error

	pageIDString := ctx.Query("page_id")
	if pageIDString != "" {
		pageID, err = strconv.Atoi(pageIDString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse page_id"})
			return
		}
	}

	if pageID < 1 {
		pageID = 1
	}

	var param = struct {
		Offset int
		Limit  int
	}{
		Offset: pageSize * (pageID - 1),
		Limit:  pageSize,
	}

	cartItems, length, err := repository.GetAllCartItems(collection, param.Offset, param.Limit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusInternalServerError, gin.H{"not found": "No product in your cart matches the params specified"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := GetAllUserCartResult{
		PageID:        pageID,
		NumberOfPages: int(math.Ceil(float64(length) / float64(pageSize))),
		ResultsFound:  int(length),
		Data:          cartItems,
	}

	if response.NumberOfPages < 1 {
		response.PageID = 0
	}

	ctx.JSON(http.StatusOK, response)
}

// DeleteCartItem is an admin specific handler to delete a single item in cart
func (a *AdminController) DeleteCartItem(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "cart")

	cartItemID := ctx.Param("id")
	if cartItemID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid param - no id specified"})
		return
	}

	_, err := repository.DeleteCartItem(collection, cartItemID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("deleted cart item with id - %s", cartItemID)
	ctx.JSON(http.StatusOK, gin.H{"success": response})
}

// DeleteAllUserCartItems is an admin specific handler to delete all items in cart of a user
func (a *AdminController) DeleteAllUserCartItems(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "cart")

	username := ctx.Param("username")
	if username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid param - no id specified"})
		return
	}

	result, err := repository.DeleteAllUserCartItems(collection, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("deleted %d cart items", result.DeletedCount)
	ctx.JSON(http.StatusOK, gin.H{"success": response})
}

// DeleteAllCartItems is an admin specific handler to delete all items in cart of different users
func (a *AdminController) DeleteAllCartItems(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "cart")

	result, err := repository.DeleteAllCartItems(collection)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("deleted %d cart items", result.DeletedCount)
	ctx.JSON(http.StatusOK, gin.H{"success": response})
}
