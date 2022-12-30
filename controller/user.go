package controller

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	auth "github.com/Emmrys-Jay/ecommerce-api/auth/jwt"
	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/entity"
	"github.com/Emmrys-Jay/ecommerce-api/repository"
	"github.com/Emmrys-Jay/ecommerce-api/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserController struct {
	*mongo.Database
}

func NewUserController(db *mongo.Database) *UserController {
	return &UserController{
		Database: db,
	}
}

// CreateUser handles requests to create a new user from a client
func (u *UserController) CreateUser(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")

	var req entity.CreateUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	user := entity.User{
		Username:       req.Username,
		Email:          req.Email,
		Fullname:       req.Fullname,
		MobileNumber:   req.MobileNumber,
		ID:             primitive.NewObjectIDFromTimestamp(time.Now()).Hex(),
		PasswordSalt:   util.RandomString(),
		EmailIsVerfied: false,
		CreatedAt:      time.Now(),
	}

	user.Password, _ = util.HashPassword(user.PasswordSalt + req.Password)

	_, err := repository.CreateUser(collection, user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	tokenMaker, err := auth.NewTokenMaker()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	token, err := tokenMaker.CreateToken(user.Username, user.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	response := entity.UserResponse{
		ID:             user.ID,
		Username:       user.Username,
		Fullname:       user.Fullname,
		Email:          user.Email,
		Token:          token,
		CreatedAt:      user.CreatedAt,
		EmailIsVerfied: user.EmailIsVerfied,
		MobileNumber:   user.MobileNumber,
	}

	ctx.JSON(http.StatusOK, response)
}

// LoginUser handles requests to confirm a user details and return a JWT token
func (u *UserController) LoginUser(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")

	var user struct {
		Username string `json:"username" form:"username" binding:"required"`
		Password string `json:"password" form:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	storedUser, err := repository.GetUser(collection, "", user.Username)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}

	if !util.PasswordIsVerified(storedUser.PasswordSalt+user.Password, storedUser.Password) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect password"})
		return
	}

	tokenMaker, err := auth.NewTokenMaker()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	token, err := tokenMaker.CreateToken(storedUser.Username, storedUser.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	response := entity.UserResponse{
		ID:             storedUser.ID,
		Username:       storedUser.Username,
		Fullname:       storedUser.Fullname,
		Email:          storedUser.Email,
		Token:          token,
		CreatedAt:      storedUser.CreatedAt,
		EmailIsVerfied: storedUser.EmailIsVerfied,
		MobileNumber:   storedUser.MobileNumber,
	}

	ctx.JSON(http.StatusOK, response)
}

// GetUser handles an admin request to get a single user stored in the database
func (u *UserController) GetUser(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")

	userID, err := util.UserIDFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	user, err := repository.GetUser(collection, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// ChangePassword handles a change password request from a user
func (u *UserController) ChangePassword(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")

	var req struct {
		Password    string `json:"password" form:"password" binding:"required"`
		NewPassword string `json:"new_password" form:"new_password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	userID, err := util.UserIDFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	user, err := repository.GetUser(collection, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	if !util.PasswordIsVerified(user.PasswordSalt+req.Password, user.Password) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"unauthorized": "incorrect password"})
		return
	}

	if req.NewPassword == req.Password {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No new password specified"})
		return
	}

	newPasswordSalt := util.RandomString()
	newHashPassword, err := util.HashPassword(newPasswordSalt + req.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	err = repository.UpdateUserFlexible(collection, user.Username, "password", newHashPassword, newPasswordSalt)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"response": "Password successfully changed"})
}

type UpdateUserRequest struct {
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
func (u *UserController) UpdateUserFlexible(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")
	var req UpdateUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	if req.Detail == "" || req.Update == "" {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(errors.New("Invalid params")))
		return
	}

	if req.Detail == "password" {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(errors.New("Invalid params")))
		return
	}

	userID, err := util.UserIDFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	err = repository.UpdateUserFlexible(collection, userID, req.Detail, req.Update, "")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"response": fmt.Sprintf("%s successfully changed", req.Detail)})
}

// AddLocation handles a register/add location request from a users account
func (u *UserController) AddLocation(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")
	var req entity.Location

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	userID, err := util.UserIDFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	_, err = repository.AddLocation(collection, userID, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	var response = struct {
		Response string          `json:"response"`
		Data     entity.Location `json:"data"`
	}{
		Response: "successfully added location",
		Data:     req,
	}
	ctx.JSON(http.StatusOK, response)
}

// VerifyEmail function
