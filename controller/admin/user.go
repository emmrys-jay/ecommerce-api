package controller

import (
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
	"go.mongodb.org/mongo-driver/mongo"
)

type AdminController struct {
	*controller.UserController
}

func NewAdminController(userController *controller.UserController) *AdminController {
	return &AdminController{
		UserController: userController,
	}
}

// GetUser handles an admin request to get a single user stored in the database
func (a *AdminController) GetUser(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "users")
	username := ctx.Param("username")

	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid params"})
		return
	}

	user, err := repository.GetUser(collection, username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, err)
			return
		}
		ctx.JSON(http.StatusInternalServerError, err)
	}

	response := controller.GetUserResponse{
		ID:                      user.ID,
		Username:                user.Username,
		Fullname:                user.Fullname,
		Email:                   user.Email,
		EmailIsVerfied:          user.EmailIsVerfied,
		MobileNumber:            user.MobileNumber,
		DefaultPaymentMethod:    user.DefaultPaymentMethod,
		SavedPaymentDetails:     user.SavedPaymentDetails,
		DefaultDeliveryLocation: user.DefaultDeliveryLocation,
		CreatedAt:               user.CreatedAt,
	}

	ctx.JSON(http.StatusOK, response)
}

type GetAllUsersResult struct {
	PageID        int           `json:"page_id"`
	ResultsFound  int           `json:"results_found"`
	NumberOfPages int           `json:"no_of_pages"`
	Data          []entity.User `json:"data"`
}

// GetAllUsers handles an admin request to get all users stored in a database
func (a *AdminController) GetAllUsers(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "users")
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
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, err)
			return
		}
		ctx.JSON(http.StatusInternalServerError, err)
	}

	response := GetAllUsersResult{
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
	Username string `json:"username" form:"username" binding:"required"`
	Detail   string `json:"detail" form:"detail" binding:"required"` //field to be updated
	Update   string `json:"update" form:"update" binding:"required"`
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
	collection := db.GetCollection(a.Database, "users")
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

	err := repository.UpdateUserFlexible(collection, req.Username, req.Detail, req.Update, "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("%s successfully changed", req.Detail)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
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
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

// DeleteUser handles a delete all users request from an admin account
func (a *AdminController) DeleteAllUsers(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "users")

	result, err := repository.DeleteAllUsers(collection)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("successfully deleted %d users", result.DeletedCount)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

type AddLocationRequest struct {
	Username string          `json:"username" form:"username" binding:"required"`
	Location entity.Location `json:"location" binding:"required"`
}

// type Location struct {
// 	HouseNumber string `json:"house_number,omitempty"`
// 	PhoneNo     string `json:"telephone,omitempty"`
// 	Street      string `json:"street,omitempty"`
// 	CityOrTown  string `json:"city_or_town,omitempty"`
// 	State       string `json:"state,omitempty"`
// 	Country     string `json:"country,omitempty"`
// 	ZipCode     string `json:"zip_code,omitempty"`
// }

// AddLocation handles a register/add location request from a users account
func (a *AdminController) AddLocation(ctx *gin.Context) {
	collection := db.GetCollection(a.Database, "users")
	var req AddLocationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	location, err := repository.AddLocation(collection, req.Username, req.Location)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	var response = struct {
		Response string           `json:"response"`
		Data     *entity.Location `json:"data"`
	}{
		Response: "successfully added location",
		Data:     location,
	}
	ctx.JSON(http.StatusOK, response)
}
