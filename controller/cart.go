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

type AddToCartRequest struct {
	ProductID string `json:"product_id" form:"product_id"`
	Quantity  int64  `json:"quantity" form:"quantity,min=1"`
}

// AddToCart response to add to cart requests from a user
func (u *UserController) AddToCart(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "cart")
	var req AddToCartRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	if req.ProductID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params"})
		return
	}

	if req.Quantity == 0 {
		req.Quantity = 1
	}

	username, err := util.UsernameFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	result, err := repository.AddToCart(collection, req.Quantity, req.ProductID, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("added %d products successfully to cart with id: %s", req.Quantity, result.InsertedID)
	ctx.JSON(http.StatusOK, gin.H{"result": response})
}

// RemoveFromCart totally removes a product associated with a particular user from cart
func (u *UserController) RemoveFromCart(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "cart")

	cartItemID := ctx.Param("cart-id")

	if cartItemID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params"})
		return
	}

	username, err := util.UsernameFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	result, err := repository.RemoveFromCart(collection, cartItemID, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	if result.DeletedCount == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"not found": "specified params did not match any document"})

	}

	ctx.JSON(http.StatusOK, gin.H{"success": "removed product from cart"})
}

type UpdateCartRequest struct {
	CartID   string `json:"cart_id" form:"cart_id"`
	Quantity int    `json:"quantity" form:"quantity,min=1"`
}

// UpdateCartQuantity changes the quantity of products stored in cart
func (u *UserController) UpdateCartQuantity(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "cart")
	var req UpdateCartRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	if req.CartID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params - id"})
		return
	}

	if req.Quantity == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params - qunatity"})
		return
	}

	username, err := util.UsernameFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	_, err = repository.UpdateCartQuantity(collection, req.Quantity, req.CartID, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": "updated quantity of product in cart"})

}

type GetUserCartItemsResult struct {
	PageID        int               `json:"page_id"`
	ResultsFound  int               `json:"results_found"`
	NumberOfPages int               `json:"no_of_pages"`
	Data          []entity.CartItem `json:"data"`
}

// GetUserCartItems gets all products stored in a users cart
func (u *UserController) GetUserCartItems(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "cart")
	var pageID, pageSize = 1, 5
	var err error

	pageIDString := ctx.Query("page_id")
	if pageIDString != "" {
		pageID, err = strconv.Atoi(pageIDString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "could not parse page_id param"})
			return
		}
	}

	username, err := util.UsernameFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	var param = struct {
		Offset int
		Limit  int
	}{
		Offset: pageSize * (pageID - 1),
		Limit:  pageSize,
	}

	cartItems, length, err := repository.GetUserCartItems(collection, username, param.Offset, param.Limit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, gin.H{"not found": "No items in cart currently"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := GetUserCartItemsResult{
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

// Make ProductName and Username unique
