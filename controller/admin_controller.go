package controller

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	auth "github.com/Emmrys-Jay/ecommerce-api/auth/jwt"
	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	util "github.com/Emmrys-Jay/ecommerce-api/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type AdminController struct {
	*UserController
}

func NewAdminController(userController *UserController) *AdminController {
	return &AdminController{
		UserController: userController,
	}
}

// GetCartItem gets a single product stored in a users cart, mainly used by admin
func (a *AdminController) GetCartItem(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "cart")

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
			ctx.JSON(http.StatusInternalServerError, gin.H{"not found": "No product in your cart matches the params specified"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, cartItem)
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
	var pageSize int

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
		ctx.JSON(http.StatusBadRequest, gin.H{"response": "invalid params - page_id"})
		return
	}

	pageSizeString := ctx.Query("page_size")
	if pageSizeString == "" {
		pageSize = 10
	} else {
		pageSize, err = strconv.Atoi(pageSizeString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"response": "invalid params - page_size"})
			return
		}
	}

	if pageSize < 5 {
		ctx.JSON(http.StatusBadRequest, gin.H{"response": "invalid params - page_size"})
		return
	}

	var params = struct {
		Limit  int
		Offset int
	}{
		Limit:  pageSize,
		Offset: pageSize * (pageID - 1),
	}

	orders, err := repository.GetAllOrders(collection, params.Limit, params.Offset)
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
		NumberOfPages: int(math.Ceil(float64(len(orders)) / float64(pageSize))),
		ResultsFound:  len(orders),
		Data:          orders,
	}

	ctx.JSON(http.StatusOK, response)

}

// DeleteProductRequest stores delete request params
type DeleteProductRequest struct {
	Name string `json:"name" form:"name" bson:"name"`
}

// DeleteProduct deletes a document whose exact name is specified
func (a *AdminController) DeleteProduct(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "products")
	var req DeleteProductRequest // create variable to be used to send data to db package

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	result, err := repository.DeleteProduct(collection, req.Name)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}
	response := fmt.Sprintf("deleted %d document with name: %s", result.DeletedCount, req.Name)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

// DeleteAllProducts deletes all documents in the products collection
func (a *AdminController) DeleteAllProducts(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "products")

	result, err := repository.DeleteAllProducts(collection)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("deleted %d document(s)", result.DeletedCount)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

// UpdateProductRequest stores update request params
type UpdateProductRequest struct {
	Name     string           `json:"name" form:"name" binding:"required"`
	Price    float64          `json:"price" form:"price"`
	Quantity int64            `json:"quantity" form:"quantity"`
	Features []entity.Feature `json:"features" form:"features"`
}

// UpdateProduct updates either the price or quantity of a product specified with its exact name
func (a *AdminController) UpdateProduct(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "products")
	var req UpdateProductRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	if req.Name == "" {
		ctx.JSON(http.StatusNotAcceptable, gin.H{"error": "nothing specified to update"})
		return
	}

	if req.Price == 0 && req.Quantity == 0 && req.Features == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid update params"})
		return
	}

	result, err := repository.UpdateProduct(collection, req.Name, req.Price, req.Quantity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	response := fmt.Sprintf("updated %d document with name: %s", result.ModifiedCount, req.Name)
	ctx.JSON(http.StatusOK, gin.H{"response": response})

}

// GetAllUsers handles an admin request to get all users stored in a database
func (a *AdminController) GetAllUsers(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "users")

	users, err := repository.GetAllUsers(collection)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, err)
			return
		}
		ctx.JSON(http.StatusInternalServerError, err)
	}

	ctx.JSON(http.StatusOK, users)
}

// DeleteUser handles a delete user request from an admin account
func (a *AdminController) DeleteUser(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "users")

	username := ctx.Query("username")

	_, err := repository.DeleteUser(collection, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("successfully deleted user with username: %s", username)
	ctx.JSON(http.StatusUnauthorized, gin.H{"response": response})
}

// DeleteUser handles a delete all users request from an admin account
func (a *AdminController) DeleteAllUsers(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "users")

	result, err := repository.DeleteAllUsers(collection)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("successfully deleted %d users", result)
	ctx.JSON(http.StatusUnauthorized, gin.H{"response": response})
}
