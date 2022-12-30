package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	util "github.com/Emmrys-Jay/ecommerce-api/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
*  AddProducts adds a single documents from the request body to the database
*	Required fields:
*	- Name
*	- Price
*	- Currency
*	- Quantity
*	- Description
*	- Category
*```
*	Other Fields:
*	- Features: Slice of Feature Object
*   - SlashedPrice
*   - Pictures: Slice of String
*   - Videos: Slice of String
 */
func (a *AdminController) AddOneProduct(ctx *gin.Context) {
	collection := db.GetCollection(a.UserController.Database, "products")
	var req = entity.Product{}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	// Assign time values to the time fields before sending to the database
	req.ID = primitive.NewObjectIDFromTimestamp(time.Now()).Hex()
	req.CreatedAt = time.Now()
	req.LastUpdated = time.Now()
	req.NumOfOrders = 0

	_, err := repository.InsertOneProduct(collection, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, bson.M{"result": "added successfully"})
}

// AddProducts adds multiple documents from the request body to the database
func (a *AdminController) AddProducts(ctx *gin.Context) {
	collection := db.GetCollection(a.UserController.Database, "products")
	var req = []entity.Product{}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	currentTime := time.Now()

	// Assign time values to the time fields before sending to the database
	for i := range req {
		req[i].ID = primitive.NewObjectIDFromTimestamp(time.Now()).Hex()
		req[i].CreatedAt = currentTime
		req[i].LastUpdated = currentTime
		req[i].NumOfOrders = 0

	}

	result, err := repository.InsertProducts(collection, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("added %d products successfully", len(result.InsertedIDs))
	ctx.JSON(http.StatusOK, gin.H{"result": response})
}

// DeleteProductRequest stores delete request params
type DeleteProductRequest struct {
	Name string `json:"name" form:"name" bson:"name"`
}

// DeleteProduct deletes a document whose exact name is specified
func (a *AdminController) DeleteProducts(ctx *gin.Context) {
	collection := db.GetCollection(a.UserController.Database, "products")

	ids := ctx.Query("id")
	if ids == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "product(s) not specified"})
		return
	}

	length, err := repository.DeleteProduct(collection, ids)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}
	response := fmt.Sprintf("deleted %d document(s)", length)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

// DeleteAllProducts deletes all documents in the products collection
func (a *AdminController) DeleteAllProducts(ctx *gin.Context) {
	collection := db.GetCollection(a.UserController.Database, "products")

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
	Price    float64 `json:"price" form:"price"`
	Quantity int64   `json:"quantity" form:"quantity"`
}

// UpdateProduct updates either the price or quantity of a product specified with its id
func (a *AdminController) UpdateProduct(ctx *gin.Context) {
	collection := db.GetCollection(a.UserController.Database, "products")
	var req UpdateProductRequest

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusNotAcceptable, gin.H{"error": "nothing specified to update"})
		return
	}

	if req.Price == 0 && req.Quantity == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid update params"})
		return
	}

	result, err := repository.UpdateProduct(collection, id, req.Price, req.Quantity, 0)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	response := fmt.Sprintf("updated %d document with id: %s", result.ModifiedCount, id)
	ctx.JSON(http.StatusOK, gin.H{"response": response})

}
