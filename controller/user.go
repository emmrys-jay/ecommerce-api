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
	ID             string    `json:"_id"`
	Username       string    `json:"username"`
	Fullname       string    `json:"fullname"`
	Email          string    `json:"email"`
	Token          string    `json:"token"`
	CreatedAt      time.Time `json:"created_at"`
	EmailIsVerfied bool      `json:"email_is_verified"`
	MobileNumber   string    `json:"mobile_number,omitempty"`
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

	user.ID = primitive.NewObjectIDFromTimestamp(time.Now()).String()[10:34]
	user.PasswordSalt = util.RandomString()
	user.HashedPassword, _ = util.HashPassword(user.PasswordSalt + user.HashedPassword)
	user.EmailIsVerfied = false
	user.CreatedAt = time.Now()

	_, err := repository.CreateUser(collection, user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	storedUser, err := repository.GetUser(collection, user.ID)
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

// LoginUser handles requests to confirm a user details and return a JWT token
func (u *UserController) LoginUser(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")
	var user LoginUserRequest

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	storedUser, err := repository.GetUser(collection, "", user.Username)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
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

type GetUserResponse struct {
	ID                      string            `json:"_id,omitempty"`
	Username                string            `json:"username,omitempty"`
	Fullname                string            `json:"fullname,omitempty"`
	Email                   string            `json:"email,omitempty"`
	EmailIsVerfied          bool              `json:"email_is_verified,omitempty"`
	MobileNumber            string            `json:"mobile_number,omitempty"`
	ProfilePicture          string            `json:"picture,omitempty" bson:"picture"`
	DefaultPaymentMethod    string            `json:"default_payment_method,omitempty"`
	SavedPaymentDetails     string            `json:"saved_payment_details,omitempty"`
	DefaultDeliveryLocation entity.Location   `json:"default_delivery_location,omitempty"`
	RegisteredLocations     []entity.Location `json:"registered_locations,omitempty"`
	FavouriteProducts       []entity.Product  `json:"favourite_products,omitempty" bson:"favourite_products"`
	Orders                  []entity.Order    `json:"orders,omitempty" bson:"orders"`
	CreatedAt               time.Time         `json:"created_at,omitempty"`
	LastUpdated             time.Time         `json:"last_updated,omitempty" bson:"last_updated"`
}

// GetUser handles an admin request to get a single user stored in the database
func (u *UserController) GetUser(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")

	userID, err := util.UserIDFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	user, err := repository.GetUser(collection, userID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, err)
			return
		}
		ctx.JSON(http.StatusInternalServerError, err)
	}

	response := GetUserResponse{
		ID:             user.ID,
		Username:       user.Username,
		Fullname:       user.Fullname,
		Email:          user.Email,
		EmailIsVerfied: user.EmailIsVerfied,
		MobileNumber:   user.MobileNumber,
		ProfilePicture: user.ProfilePicture,

		DefaultPaymentMethod:    user.DefaultPaymentMethod,
		SavedPaymentDetails:     user.SavedPaymentDetails,
		DefaultDeliveryLocation: user.DefaultDeliveryLocation,
		RegisteredLocations:     user.RegisteredLocations,
		CreatedAt:               user.CreatedAt,
		LastUpdated:             user.LastUpdated,
		Orders:                  user.Orders,
		FavouriteProducts:       user.FavouriteProducts,
	}

	ctx.JSON(http.StatusOK, response)
}

type ChangePasswordRequest struct {
	// Username    string `json:"username" form:"password" binding:"required"`
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

	userID, err := util.UserIDFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	user, err := repository.GetUser(collection, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	if !util.PasswordIsVerified(user.PasswordSalt+req.Password, user.HashedPassword) {
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
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	err = repository.UpdateUserFlexible(collection, user.Username, "password", newHashPassword, newPasswordSalt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"response": "Password successfully changed"})
}

type UpdateUserRequest struct {
	// Username string `json:"username" form:"username" binding:"required"`
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params"})
		return
	}

	if req.Detail == "password" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params"})
		return
	}

	username, err := util.UsernameFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	err = repository.UpdateUserFlexible(collection, username, req.Detail, req.Update, "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(err))
		return
	}

	response := fmt.Sprintf("%s successfully changed", req.Detail)
	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

type AddLocationRequest struct {
	// Username string          `json:"username" form:"username" binding:"required"`
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
func (u *UserController) AddLocation(ctx *gin.Context) {
	collection := db.GetCollection(u.Database, "users")
	var req AddLocationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(err))
		return
	}

	userID, err := util.UserIDFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not get logged in user from token"})
		return
	}

	location, err := repository.AddLocation(collection, userID, req.Location)
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
