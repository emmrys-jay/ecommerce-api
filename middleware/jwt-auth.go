package middleware

import (
	"net/http"
	"strings"

	"github.com/Emmrys-Jay/ecommerce-api/token"
	"github.com/gin-gonic/gin"
)

func AuthorizeJWT() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth := ctx.GetHeader("Authorization")
		if auth == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"unauthorized": "access denied"})
			ctx.Abort()
			return
		}

		jwtToken := strings.Split(auth, " ")[1]
		tokenMaker, _ := token.NewTokenMaker()

		payload, err := tokenMaker.VerifyToken(jwtToken)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"unauthorized": "access denied"})
			ctx.Abort()
			return
		}

		if err := payload.Valid(); err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"unauthorized": "access denied"})
			ctx.Abort()
			return
		}
	}
}
