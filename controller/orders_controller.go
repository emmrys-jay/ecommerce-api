package controller

import (
	"fmt"
	"net/http"
	"strings"

	auth "github.com/Emmrys-Jay/ecommerce-api/auth/jwt"
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
	ProductName   string          `json:"product_name" binding:"required"`
	Quantity      int             `json:"quantity"`
	Location      entity.Location `json:"location" binding:"required"`
	PaymentMethod string          `json:"payment_method" binding:"required"`
}

// OrderProduct serves an order a product request from a user directly
func (u *UserController) OrderProduct(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "orders")
	var req OrderProductRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	tokenMaker, _ := auth.NewTokenMaker()

	// Get Authorization header and split it to get JWT token
	// Verify JWT token to get custom payload which contains username information
	tokenString := strings.Split(ctx.GetHeader("Authorization"), " ")[1]
	payload, _ := tokenMaker.VerifyToken(tokenString)

	result, err := repository.OrderProductDirectly(collection, req.Location, req.Quantity, payload.Username, req.Fullname, req.ProductName, req.PaymentMethod)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	_, err = repository.UpdateProduct(db.GetCollection(u.Database, "products"), req.ProductName, 0.00, -int64(req.Quantity))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	order, err := repository.GetSingleOrder(collection, payload.Username, result.InsertedID.(primitive.ObjectID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	// Add order to the orders in the user collection
	_, err = repository.AddOrderToUser(collection, payload.Username, order)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("you just bought %d product(s) with name - %s and id - %s", req.Quantity, req.ProductName, result.InsertedID)
	ctx.JSON(http.StatusOK, gin.H{"success": response})
}

// GetOrdersWithUsername gets orders associated with a username
func (u *UserController) GetOrdersWithUsername(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "orders")

	username := ctx.Query("username")
	if username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"response": "invalid params"})
		return
	}

	orders, err := repository.GetOrdersWithUsername(collection, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

// type DeliverOrderRequest struct {
// 	Username string             `json:"username" binding:"required"`
// 	OrderID  primitive.ObjectID `json:"id" binding:"required"`
// }

// func (server *Server) DeliverOrder(ctx gin.Context) {
// 	collection := db.GetCollection(server.DB, "orders")
// 	var req DeliverOrderRequest

// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"response": "invalid params"})
// 		return
// 	}

// 	result, err := repository.DeliverOrder(collection, req.Username, req.OrderID)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
// 		return
// 	}

// 	order, err := repository.GetSingleOrder(collection, req.Username, req.OrderID)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
// 		return
// 	}

// }

// Fix Issue with the number of results found for a GetAllUser search
