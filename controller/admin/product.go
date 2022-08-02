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
	collection := db.GetCollection(a.Database, "products")
	var req = entity.Product{}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	// Assign time values to the time fields before sending to the database
	req.ID = primitive.NewObjectIDFromTimestamp(time.Now()).String()
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
	collection := db.GetCollection(a.Database, "products")
	var req = []entity.Product{}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	currentTime := time.Now()

	// Assign time values to the time fields before sending to the database
	for i := range req {
		req[i].ID = primitive.NewObjectIDFromTimestamp(time.Now()).String()[10:34]
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
	ID       string           `json:"id" form:"name" binding:"required"`
	Price    float64          `json:"price" form:"price"`
	Quantity int64            `json:"quantity" form:"quantity"`
	Features []entity.Feature `json:"features" form:"features"`
}

// UpdateProduct updates either the price or quantity of a product specified with its id
func (a *AdminController) UpdateProduct(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "products")
	var req UpdateProductRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	if req.ID == "" {
		ctx.JSON(http.StatusNotAcceptable, gin.H{"error": "nothing specified to update"})
		return
	}

	if req.Price == 0 && req.Quantity == 0 && req.Features == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid update params"})
		return
	}

	result, err := repository.UpdateProduct(collection, req.ID, req.Price, req.Quantity, req.Quantity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	response := fmt.Sprintf("updated %d document with id: %s", result.ModifiedCount, req.ID)
	ctx.JSON(http.StatusOK, gin.H{"response": response})

}
