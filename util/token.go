package util

import (
	"strings"

	auth "github.com/Emmrys-Jay/ecommerce-api/auth/jwt"
	"github.com/gin-gonic/gin"
)

func UsernameFromToken(ctx *gin.Context) (string, error) {
	tokenMaker, err := auth.NewTokenMaker()
	if err != nil {
		return "", err
	}

	// Get Authorization header and split it to get JWT token
	// Verify JWT token to get custom payload which contains username information
	tokenString := strings.Split(ctx.GetHeader("Authorization"), " ")[1]
	payload, err := tokenMaker.VerifyToken(tokenString)
	if err != nil {
		return "", err
	}

	return payload.Username, nil
}

func UserIDFromToken(ctx *gin.Context) (string, error) {
	tokenMaker, err := auth.NewTokenMaker()
	if err != nil {
		return "", err
	}

	// Get Authorization header and split it to get JWT token
	// Verify JWT token to get custom payload which contains username information
	tokenString := strings.Split(ctx.GetHeader("Authorization"), " ")[1]
	payload, err := tokenMaker.VerifyToken(tokenString)
	if err != nil {
		return "", err
	}

	return payload.ID, nil
}
