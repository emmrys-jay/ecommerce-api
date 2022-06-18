package handlers

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddProducts adds a single documents from the request body to the database
func (server *Server) AddOneProduct(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "products")
	var req = db.Product{}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Assign time values to the time fields before sending to the database
	req.CreatedAt = time.Now()
	req.LastUpdated = time.Now()

	_, err := db.InsertOneProduct(collection, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, bson.M{"result": "added successfully"})
}

// AddProducts adds multiple documents from the request body to the database
func (server *Server) AddProducts(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "products")
	var req = []db.Product{}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Assign time values to the time fields before sending to the database
	for _, val := range req {
		val.CreatedAt = time.Now()
		val.LastUpdated = time.Now()
	}

	result, err := db.InsertProducts(collection, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	response := fmt.Sprintf("added %d products successfully", len(result.InsertedIDs))
	ctx.JSON(http.StatusOK, gin.H{"result": response})
}

// FindProductsRequest stores find products request params
type FindProductsRequest struct {
	Name   string `form:"name" bson:"name"`
	PageID int64  `form:"page_id" bson:"page_id"`
}

// FindProductsResult models the find products request result
type FindProductsResult struct {
	PageID        int64        `json:"page_id" bson:"page_id"`
	ResultsFound  int64        `json:"results_found"`
	NumberOfPages int64        `json:"no_of_pages"`
	Data          []db.Product `json:"data"`
}

// FindProducts returns the documents that match the regex pattern sent into the database
func (server *Server) FindProducts(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "products")
	var req FindProductsRequest
	var pageSize int64 = 5
	req.PageID = 1

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if req.PageID < 1 {
		req.PageID = 1
	}

	// initialize and define new struct with names easy to understand while querying the database
	var param = struct {
		Name   string
		Offset int64
		Limit  int64
	}{
		Name:   req.Name,
		Offset: pageSize * (req.PageID - 1),
		Limit:  pageSize,
	}

	products, err := db.FindProducts(collection, param.Name, param.Offset, param.Limit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	NoOfPages := float64(len(products)) / float64(int(pageSize))

	result := FindProductsResult{
		PageID:        req.PageID,
		ResultsFound:  int64(len(products)),
		NumberOfPages: int64(math.Ceil(NoOfPages)),
		Data:          products,
	}

	ctx.JSON(http.StatusOK, result)
}

// DeleteProductRequest stores delete request params
type DeleteProductRequest struct {
	Name string `form:"name" bson:"name"`
}

// DeleteProduct deletes a document whose exact name is specified
func (server *Server) DeleteProduct(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "products")
	var req DeleteProductRequest // create variable to be used to send data to db package

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	result, err := db.DeleteProduct(collection, req.Name)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	response := fmt.Sprintf("deleted %d document with name: %s", result.DeletedCount, req.Name)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

// DeleteAllProducts deletes all documents in the products collection
func (server *Server) DeleteAllProducts(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "products")

	result, err := db.DeleteAllProducts(collection)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := fmt.Sprintf("deleted %d document(s)", result.DeletedCount)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

// UpdateProductRequest stores update request params
type UpdateProductRequest struct {
	Name     string  `form:"name"`
	Price    float64 `form:"price"`
	Quantity int64   `form:"quantity"`
}

// UpdateProduct updates either the price or quantity of a product specified with its exact name
func (server *Server) UpdateProduct(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "products")
	var req UpdateProductRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if req.Name == "" {
		ctx.JSON(http.StatusNotAcceptable, gin.H{"error": "nothing specified to update"})
		return
	}

	if req.Price == 0 && req.Quantity == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "nothing specified to update"})
		return
	}

	result, err := db.UpdateProduct(collection, req.Name, req.Price, req.Quantity)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	response := fmt.Sprintf("updated %d document with name: %s", result.ModifiedCount, req.Name)
	ctx.JSON(http.StatusOK, gin.H{"response": response})

}

// AddReviewRequest stores the request value for add review request
type AddReviewRequest struct {
	Name string `form:"name" binding:"required"`
}

// AddReview adds a review to a single product
func (server *Server) AddReview(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "products")
	var review db.Review
	var name AddReviewRequest

	if err := ctx.ShouldBindJSON(&review); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindQuery(&name); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	result, err := db.AddProductReview(collection, name.Name, review)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := fmt.Sprintf("updated %d document with name: %s", result.ModifiedCount, name.Name)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}
