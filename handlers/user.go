package handlers

import (
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
	ID        primitive.ObjectID `json:"_id"`
	Username  string             `json:"username" binding:"required"`
	Fullname  string             `json:"fullname" binding:"required"`
	Email     string             `json:"email" binding:"email,required"`
	Token     string             `json:"token"`
	CreatedAt time.Time          `json:"created_at" binding:"required"`
}

// CreateUser handles requests to create a new user from a client
func (server Server) CreateUser(ctx *gin.Context) {
	collection := db.GetCollection(server.DB, "users")

	var user db.CreateUserRequest
	var response UserResponse

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user.Password, _ = util.HashPassword(user.Password)

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

	response = UserResponse{
		ID:        storedUser.ID,
		Username:  storedUser.Username,
		Fullname:  storedUser.Fullname,
		Email:     storedUser.Email,
		Token:     token,
		CreatedAt: storedUser.CreatedAt,
	}

	ctx.JSON(http.StatusOK, response)
}

// LoginUserRequest models the login user request json body
type LoginUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login user handles requests to confirm a user details and return a JWT token
func (server Server) LoginUser(ctx *gin.Context) {
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
		ID:        storedUser.ID,
		Username:  storedUser.Username,
		Fullname:  storedUser.Fullname,
		Email:     storedUser.Email,
		Token:     token,
		CreatedAt: storedUser.CreatedAt,
	}

	ctx.JSON(http.StatusOK, response)
}

// GetAllUsers handles an admin request to get all users stored in a database
func (server Server) GetAllUsers(ctx *gin.Context) {
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
