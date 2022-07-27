package controller

import (
	"fmt"
	"net/http"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	util "github.com/Emmrys-Jay/ecommerce-api/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type AddToCartRequest struct {
	ProductID string `json:"product_id" form:"product_name"`
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

// DecrementCartQuantity substracts one from the quantity of a particular product stored in cart
func (u *UserController) DecrementCartQuantity(ctx *gin.Context) {
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

	cartItem, err := repository.GetCartItem(collection, cartItemID, username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, gin.H{"not found": "No product in your cart matches the params specified"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	if cartItem.Quantity == 1 {
		_, err := repository.RemoveFromCart(collection, cartItemID, username)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"success": "removed product from cart"})
		return
	}

	_, err = repository.SubtractCartQuantity(collection, cartItemID, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": "subtracted quantity of product"})

}

// GetUserCartItems gets all products stored in a users cart
func (u *UserController) GetUserCartItems(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "cart")

	username, err := util.UsernameFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	cartItems, err := repository.GetUserCartItems(collection, username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, gin.H{"not found": "No items in cart currently"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, cartItems)
}

// Make ProductName and Username unique
