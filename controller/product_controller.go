package controller

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	"github.com/Emmrys-Jay/ecommerce-api/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// FindProductsRequest models find products request params
type FindProductsRequest struct {
	Name     string `json:"name" form:"name" bson:"name"`
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

	products, length, err := repository.FindProducts(collection, param.Name, param.Offset, param.Limit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	NoOfPages := math.Ceil(float64(length) / float64(int(req.PageSize)))

	result := FindProductsResult{
		PageID:        req.PageID,
		ResultsFound:  int64(length),
		NumberOfPages: int64(NoOfPages),
		Data:          products,
	}

	ctx.JSON(http.StatusOK, result)
}

/*
*	AddReview adds a review to a single product
*	Params:
*		- Name of Product - URL param
*		- Product Review - json Body
 */

//  type Review struct {
// 	User      string    `json:"user"`
// 	Stars     int64     `json:"stars" binding:"required"`
// 	Comment   string    `json:"comment,omitempty"`
// 	CreatedAt time.Time `json:"created_at"`
// }

func (u *UserController) AddReview(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "products")
	var review entity.Review

	if err := ctx.ShouldBindJSON(&review); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	username, err := util.UsernameFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	review.User = username

	productID := ctx.Param("productID")
	if productID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid param - product ID"})
		return
	}

	result, err := repository.AddProductReview(collection, productID, review)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, util.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("updated %d document with id: %s", result.ModifiedCount, productID)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

// GetProductCategories returns the unique categories of products currently stored in the database
// func (u *UserController) GetProductCategories(ctx *gin.Context) {
// 	collection := db.GetCollection(u.Database, "products")

// }

// GetProductByCategory returns products that belong to a category
func (u *UserController) GetProductsByCategory(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "products")
	pageSize, pageID := 5, 1
	var err error

	ctgy := ctx.Param("category") // Get category from url path
	pageIDString := ctx.Query("page_id")

	if ctgy == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid url param"})
		return
	}

	if pageIDString != "" {
		pageID, err = strconv.Atoi(pageIDString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
			return
		}

		if pageID < 1 {
			pageID = 1
		}
	}

	var param = struct {
		Offset int
		Limit  int
	}{
		Offset: pageSize * (pageID - 1),
		Limit:  pageSize,
	}

	products, length, err := repository.GetProductsByCategory(collection, ctgy, param.Offset, param.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	NoOfPages := math.Ceil(float64(length) / float64(int(pageSize)))

	result := FindProductsResult{
		PageID:        int64(pageID),
		ResultsFound:  int64(length),
		NumberOfPages: int64(NoOfPages),
		Data:          products,
	}

	ctx.JSON(http.StatusOK, result)
}
