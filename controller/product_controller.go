package controller

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	"github.com/Emmrys-Jay/ecommerce-api/util"
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
func (u *UserController) AddOneProduct(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "products")
	var req = entity.Product{}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	// Assign time values to the time fields before sending to the database
	req.ID = primitive.NewObjectIDFromTimestamp(time.Now())
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
func (u *UserController) AddProducts(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "products")
	var req = []entity.Product{}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	// Assign time values to the time fields before sending to the database
	for _, val := range req {
		val.CreatedAt = time.Now()
		val.LastUpdated = time.Now()
		val.NumOfOrders = 0
	}

	result, err := repository.InsertProducts(collection, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}
	response := fmt.Sprintf("added %d products successfully", len(result.InsertedIDs))
	ctx.JSON(http.StatusOK, gin.H{"result": response})
}

// FindProductsRequest models find products request params
type FindProductsRequest struct {
	Name     string `json:"name" form:"name" bson:"name" binding:"required"`
	PageID   int64  `json:"page_id" form:"page_id" bson:"page_id"`
	PageSize int64  `json:"page_size" form:"page_size" bson:"page_size"`
}

// FindProductsResult models the find products request result
type FindProductsResult struct {
	PageID        int64            `json:"page_id"`
	ResultsFound  int64            `json:"results_found"`
	NumberOfPages int64            `json:"no_of_pages"`
	Data          []entity.Product `json:"data"`
}

// FindProducts returns the documents that match the regex pattern sent into the database
func (u *UserController) FindProducts(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "products")
	var req FindProductsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	if req.Name == "" {
		ctx.JSON(http.StatusBadRequest, "invalid params")
		return
	}

	if req.PageID < 1 {
		req.PageID = 1
	}

	if req.PageSize < 5 {
		req.PageSize = 5
	}

	// initialize and define new struct with names easy to understand while querying the database
	var param = struct {
		Name   string
		Offset int64
		Limit  int64
	}{
		Name:   req.Name,
		Offset: req.PageSize * (req.PageID - 1),
		Limit:  req.PageSize,
	}

	products, err := repository.FindProducts(collection, param.Name, param.Offset, param.Limit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	NoOfPages := math.Ceil(float64(len(products)) / float64(int(req.PageSize)))

	result := FindProductsResult{
		PageID:        req.PageID,
		ResultsFound:  int64(len(products)),
		NumberOfPages: int64(NoOfPages),
		Data:          products,
	}

	ctx.JSON(http.StatusOK, result)
}

// AddReviewRequest models the request structure for an add review request
type AddReviewRequest struct {
	Name string `json:"name" form:"name" binding:"required"`
}

/*
*	AddReview adds a review to a single product
*	Params:
*		- Name of Product - Query
*		- Product Review - json Body
 */
func (u *UserController) AddReview(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "products")
	var review entity.Review
	var name AddReviewRequest

	if err := ctx.ShouldBindJSON(&review); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	if err := ctx.ShouldBindQuery(&name); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	result, err := repository.AddProductReview(collection, name.Name, review)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("updated %d document with name: %s", result.ModifiedCount, name.Name)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

// Fix Issue with the number of results found for a FindProducts search
