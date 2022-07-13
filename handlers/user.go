package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Emmrys-Jay/ecommerce-api/db"
	"github.com/Emmrys-Jay/ecommerce-api/token"
	"github.com/Emmrys-Jay/ecommerce-api/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

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
func (server *Server) CreateUser(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "users")

	var user db.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if user.PasswordSalt != "" {
		ctx.JSON(http.StatusBadRequest, "invalid params")
		return
	}

	user.PasswordSalt = util.RandomString()
	user.HashedPassword, _ = util.HashPassword(user.PasswordSalt + user.HashedPassword)
	user.EmailIsVerfied = false
	user.CreatedAt = time.Now()

	_, err := db.CreateUser(collection, user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	storedUser, err := db.GetUser(collection, user.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	tokenMaker, err := token.NewTokenMaker()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	token, err := tokenMaker.CreateToken(storedUser.Username, storedUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
func (server *Server) LoginUser(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "users")
	var user LoginUserRequest

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	storedUser, err := db.GetUser(collection, user.Username)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if !util.PasswordIsVerified(storedUser.PasswordSalt+user.Password, storedUser.HashedPassword) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect password"})
		return
	}

	tokenMaker, err := token.NewTokenMaker()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	token, err := tokenMaker.CreateToken(storedUser.Username, storedUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
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
func (server *Server) GetUser(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "users")
	username := ctx.Query("username")

	if username == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid params"})
		return
	}

	user, err := db.GetUser(collection, username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, err)
			return
		}
		ctx.JSON(http.StatusInternalServerError, err)
	}

	ctx.JSON(http.StatusOK, user)
}

// GetAllUsers handles an admin request to get all users stored in a database
func (server *Server) GetAllUsers(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "users")

	users, err := db.GetAllUsers(collection)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, err)
			return
		}
		ctx.JSON(http.StatusInternalServerError, err)
	}

	ctx.JSON(http.StatusOK, users)
}

type ChangePasswordRequest struct {
	Username    string `json:"username" form:"username" binding:"required"`
	Password    string `json:"password" form:"password" binding:"required"`
	NewPassword string `json:"new_password" form:"new_password" binding:"required"`
}

// ChangePassword handles a change password request from a user
func (server *Server) ChangePassword(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "users")
	var req ChangePasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if req.NewPassword == req.Password {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No new password specified"})
		return
	}

	user, err := db.GetUser(collection, req.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if !util.PasswordIsVerified(user.PasswordSalt+req.Password, user.HashedPassword) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"unauthorized": "incorrect password"})
		return
	}

	newPasswordSalt := util.RandomString()
	newHashPassword, err := util.HashPassword(newPasswordSalt + req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = db.UpdateUserFlexible(collection, req.Username, "password", newHashPassword, newPasswordSalt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusUnauthorized, gin.H{"response": "Password successfully changed"})
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
func (server *Server) UpdateUserFlexible(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "users")
	var req UpdateUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
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

	err := db.UpdateUserFlexible(collection, req.Username, req.Detail, req.Update, "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := fmt.Sprintf("%s successfully changed", req.Detail)
	ctx.JSON(http.StatusUnauthorized, gin.H{"response": response})
}

type AddLocationRequest struct {
	Username string      `json:"username" form:"username" binding:"required"`
	Location db.Location `json:"location" binding:"required"`
}

// AddLocation handles a register/add location request from a users account
func (server *Server) AddLocation(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "users")
	var req AddLocationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := db.AddLocation(collection, req.Username, req.Location)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusUnauthorized, gin.H{"response": "successfully added location"})
}

// DeleteUser handles a delete user request from an admin account
func (server *Server) DeleteUser(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "users")

	username := ctx.Query("username")

	_, err := db.DeleteUser(collection, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := fmt.Sprintf("successfully deleted user with username: %s", username)
	ctx.JSON(http.StatusUnauthorized, gin.H{"response": response})
}

// DeleteUser handles a delete all users request from an admin account
func (server *Server) DeleteAllUsers(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "users")

	result, err := db.DeleteAllUsers(collection)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := fmt.Sprintf("successfully deleted %d users", result)
	ctx.JSON(http.StatusUnauthorized, gin.H{"response": response})
}

// VerifyEmail function
