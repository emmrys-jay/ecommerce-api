package handlers

import (
	"net/http"
	"strings"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/token"
	"github.com/gin-gonic/gin"
)

type AddToCartRequest struct {
	ProductName string `form:"p_name"`
	Quantity    int64  `form:"quantity"`
}

func (server *Server) AddToCart(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "cart")
	var req AddToCartRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if req.ProductName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "product name field value mising"})
		return
	}

	if req.Quantity == 0 {
		req.Quantity = 1
	}

	tokenMaker, _ := token.NewTokenMaker()

	// Get Authorization header and split it to get JWT token
	// Verify JWT token to get custom payload which contains username information
	tokenString := strings.Split(ctx.GetHeader("Authorization"), "")[1]
	payload, _ := tokenMaker.VerifyToken(tokenString)

	result, err := db.AddToCart(collection, req.Quantity, req.ProductName, payload.Username)

}
