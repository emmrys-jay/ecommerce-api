package api

import (
	"fmt"
	"net/http"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// AddProducts adds a single documents from the request body to the database
func (server *Server) AddOneProduct(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "products")
	var req = db.Product{}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

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

	_, err := db.InsertProducts(collection, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, bson.M{"result": "added successfully"})
}

type FindProductsRequest struct {
	Name   string `form:"name" bson:"name" json:"name"`
	PageID int64  `form:"page_id" bson:"page_id" json:"page_id"`
}

// FindProduct returns the documents that match the regex pattern sent into the database
func (server *Server) FindProducts(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "products")
	var req FindProductsRequest
	var pageSize int64 = 5
	req.PageID = 1

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
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
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, products)
}

type DeleteProductRequest struct {
	Name string `form:"name" bson:"name"`
}

// DeleteProduct deletes a document whose exact name is specified
func (server *Server) DeleteProduct(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "products")
	var req DeleteProductRequest

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

	result, err := db.UpdateProduct(collection, req.Name, req.Price, req.Quantity)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	response := fmt.Sprintf("updated %d document with name: %s", result.ModifiedCount, req.Name)
	ctx.JSON(http.StatusOK, gin.H{"response": response})

}

// TODO: Add a createdAt field to every document created
// TODO: Add a Add review function
// TODO: handle ErrNoDocument in Update and Find functions
