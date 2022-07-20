package controller

import (
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

// UserResponse models the response of a createuser or loginuser request
type UserResponse struct {
	ID             primitive.ObjectID `json:"_id"`
	Username       string             `json:"username"`
	Fullname       string             `json:"fullname"`
	Email          string             `json:"email"`
	Token          string             `json:"token"`
	CreatedAt      time.Time          `json:"created_at"`
	EmailIsVerfied bool               `json:"email_is_verified"`
	MobileNumber   string             `json:"mobile_number,omitempty"`
}

// CreateUser handles requests to create a new user from a client
func (u *UserController) CreateUser(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")

	var user entity.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	if user.PasswordSalt != "" {
		ctx.JSON(http.StatusBadRequest, "invalid params")
		return
	}

	user.ID = primitive.NewObjectIDFromTimestamp(time.Now())
	user.PasswordSalt = util.RandomString()
	user.HashedPassword, _ = util.HashPassword(user.PasswordSalt + user.HashedPassword)
	user.EmailIsVerfied = false
	user.CreatedAt = time.Now()

	_, err := repository.CreateUser(collection, user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	storedUser, err := repository.GetUser(collection, user.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	tokenMaker, err := auth.NewTokenMaker()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	token, err := tokenMaker.CreateToken(storedUser.Username, storedUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := UserResponse{
		ID:             storedUser.ID,
		Username:       storedUser.Username,
		Fullname:       storedUser.Fullname,
		Email:          storedUser.Email,
		Token:          token,
		CreatedAt:      storedUser.CreatedAt,
		EmailIsVerfied: storedUser.EmailIsVerfied,
	}

	if storedUser.MobileNumber != "" {
		response.MobileNumber = storedUser.MobileNumber
	}

	ctx.JSON(http.StatusOK, response)
}

// LoginUserRequest models the login user request json body
type LoginUserRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// Login user handles requests to confirm a user details and return a JWT token
func (u *UserController) LoginUser(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")
	var user LoginUserRequest

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	storedUser, err := repository.GetUser(collection, user.Username)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if !util.PasswordIsVerified(storedUser.PasswordSalt+user.Password, storedUser.HashedPassword) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect password"})
		return
	}

	tokenMaker, err := auth.NewTokenMaker()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	token, err := tokenMaker.CreateToken(storedUser.Username, storedUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := UserResponse{
		ID:             storedUser.ID,
		Username:       storedUser.Username,
		Fullname:       storedUser.Fullname,
		Email:          storedUser.Email,
		Token:          token,
		CreatedAt:      storedUser.CreatedAt,
		EmailIsVerfied: storedUser.EmailIsVerfied,
	}

	if storedUser.MobileNumber != "" {
		response.MobileNumber = storedUser.MobileNumber
	}

	ctx.JSON(http.StatusOK, response)
}

// GetUser handles an admin request to get a single user stored in the database
func (u *UserController) GetUser(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")
	username := ctx.Query("username")

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

	ctx.JSON(http.StatusOK, user)
}

type ChangePasswordRequest struct {
	Username    string `json:"username" form:"username" binding:"required"`
	Password    string `json:"password" form:"password" binding:"required"`
	NewPassword string `json:"new_password" form:"new_password" binding:"required"`
}

// ChangePassword handles a change password request from a user
func (u *UserController) ChangePassword(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")
	var req ChangePasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	if req.NewPassword == req.Password {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No new password specified"})
		return
	}

	user, err := repository.GetUser(collection, req.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	if !util.PasswordIsVerified(user.PasswordSalt+req.Password, user.HashedPassword) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"unauthorized": "incorrect password"})
		return
	}

	newPasswordSalt := util.RandomString()
	newHashPassword, err := util.HashPassword(newPasswordSalt + req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	err = repository.UpdateUserFlexible(collection, req.Username, "password", newHashPassword, newPasswordSalt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"response": "Password successfully changed"})
}

type UpdateUserRequest struct {
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
func (u *UserController) UpdateUserFlexible(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")
	var req UpdateUserRequest

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

type AddLocationRequest struct {
	Username string          `json:"username" form:"username" binding:"required"`
	Location entity.Location `json:"location" binding:"required"`
}

// AddLocation handles a register/add location request from a users account
func (u *UserController) AddLocation(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")
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

// VerifyEmail function
