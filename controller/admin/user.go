package controller

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/Emmrys-Jay/ecommerce-api/controller"
	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	util "github.com/Emmrys-Jay/ecommerce-api/util"
	"github.com/gin-gonic/gin"
)

type AdminController struct {
	UserController *controller.UserController
}

func NewAdminController(userController *controller.UserController) *AdminController {
	return &AdminController{
		UserController: userController,
	}
}

// GetUser handles an admin request to get a single user stored in the database
func (a *AdminController) GetUser(ctx *gin.Context) {
	collection := db.GetCollection(a.UserController.Database, "users")
	userID := ctx.Param("user-id")

	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid params"})
		return
	}

	user, err := repository.GetUser(collection, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, *user)
}

// GetAllUsers handles an admin request to get all users stored in a database
func (a *AdminController) GetAllUsers(ctx *gin.Context) {
	collection := db.GetCollection(a.UserController.Database, "users")
	var pageID int
	var err error
	var pageSize = 5

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
		pageID = 1
	}

	var params = struct {
		Limit  int
		Offset int
	}{
		Limit:  pageSize,
		Offset: pageSize * (pageID - 1),
	}

	users, length, err := repository.GetAllUsers(collection, params.Limit, params.Offset)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	response := entity.PaginationResponse{
		PageID:        pageID,
		NumberOfPages: int(math.Ceil(float64(length) / float64(pageSize))),
		ResultsFound:  int(length),
		Data:          users,
	}

	if response.NumberOfPages < 1 {
		response.PageID = 0
	}

	ctx.JSON(http.StatusOK, response)
}

type AdminUpdateUserRequest struct {
	UserID string `json:"user_id" form:"user_id" binding:"required"`
	Detail string `json:"detail" form:"detail" binding:"required"` //field to be updated
	Update string `json:"update" form:"update" binding:"required"`
}

/*
* UpdateUserFlexible handles an update user request from a user to update any of the following:
* - username
* - profile_picture
* - email
* - mobile number
* - default payment method
 */
func (a *AdminController) UpdateUserFlexible(ctx *gin.Context) {
	collection := db.GetCollection(a.UserController.Database, "users")
	var req AdminUpdateUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	if req.Detail == "" || req.Update == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params"})
		return
	}

	if req.Detail == "password" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params"})
		return
	}

	err := repository.UpdateUserFlexible(collection, req.UserID, req.Detail, req.Update, "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("%s successfully changed", req.Detail)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

// DeleteUser handles a delete user request from an admin account
func (a *AdminController) DeleteUser(ctx *gin.Context) {
	collection := db.GetCollection(a.UserController.Database, "users")

	id := ctx.Param("user-id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(errors.New("user id is not specified")))
		return
	}

	_, err := repository.DeleteUser(collection, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("successfully deleted user with id: %s", id)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

// DeleteUser handles a delete all users request from an admin account
func (a *AdminController) DeleteAllUsers(ctx *gin.Context) {
	collection := db.GetCollection(a.UserController.Database, "users")

	result, err := repository.DeleteAllUsers(collection)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("successfully deleted %d users", result.DeletedCount)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}
