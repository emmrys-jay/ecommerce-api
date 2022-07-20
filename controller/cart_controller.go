package controller

import (
	"fmt"
	"net/http"
	"strings"

	auth "github.com/Emmrys-Jay/ecommerce-api/auth/jwt"
	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	util "github.com/Emmrys-Jay/ecommerce-api/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type AddToCartRequest struct {
	ProductName string `json:"product_name" form:"product_name"`
	Quantity    int64  `json:"quantity" form:"quantity,min=1"`
}

// AddToCart response to add to cart requests from a user
func (u *UserController) AddToCart(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "cart")
	var req AddToCartRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	if req.ProductName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params"})
		return
	}

	if req.Quantity == 0 {
		req.Quantity = 1
	}

	tokenMaker, _ := auth.NewTokenMaker()

	// Get Authorization header and split it to get JWT token
	// Verify JWT token to get custom payload which contains username information
	tokenString := strings.Split(ctx.GetHeader("Authorization"), " ")[1]
	payload, _ := tokenMaker.VerifyToken(tokenString)

	result, err := repository.AddToCart(collection, req.Quantity, req.ProductName, payload.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("added %d products successfully with id: %s", req.Quantity, result.InsertedID)
	ctx.JSON(http.StatusOK, gin.H{"result": response})
}

// RemoveFromCart totally removes a product associated with a particular user from cart
func (u *UserController) RemoveFromCart(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "cart")

	productName := ctx.Query("product_name")

	if productName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params"})
		return
	}

	tokenMaker, _ := auth.NewTokenMaker()

	// Get Authorization header and split it to get JWT token
	// Verify JWT token to get custom payload which contains username information
	tokenString := strings.Split(ctx.GetHeader("Authorization"), " ")[1]
	payload, _ := tokenMaker.VerifyToken(tokenString)

	result, err := repository.RemoveFromCart(collection, productName, payload.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	if result.DeletedCount == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"not found": "specified params did not match any document"})

	}
}

// DecrementCartQuantity substracts one from the quantity of a particular product stored in cart
func (u *UserController) DecrementCartQuantity(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "cart")

	productName := ctx.Query("product_name")

	if productName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params"})
		return
	}

	tokenMaker, _ := auth.NewTokenMaker()

	// Get Authorization header and split it to get JWT token
	// Verify JWT token to get custom payload which contains username information
	tokenString := strings.Split(ctx.GetHeader("Authorization"), " ")[1]
	payload, _ := tokenMaker.VerifyToken(tokenString)

	cartItem, err := repository.GetCartItem(collection, productName, payload.Username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, gin.H{"not found": "No product in your cart matches the params specified"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	if cartItem.Quantity == 1 {
		_, err := repository.RemoveFromCart(collection, productName, payload.Username)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"success": "removed product from cart"})
		return
	}

	_, err = repository.SubtractCartQuantity(collection, productName, payload.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": "subtracted quantity of product"})

}

// GetUserCartItems gets all products stored in a users cart
func (u *UserController) GetUserCartItems(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "cart")

	tokenMaker, _ := auth.NewTokenMaker()

	// Get Authorization header and split it to get JWT token
	// Verify JWT token to get custom payload which contains username information
	tokenString := strings.Split(ctx.GetHeader("Authorization"), " ")[1]
	payload, _ := tokenMaker.VerifyToken(tokenString)

	cartItems, err := repository.GetUserCartItems(collection, payload.Username)
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
