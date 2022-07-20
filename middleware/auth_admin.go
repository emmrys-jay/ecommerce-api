package middleware

import (
	"net/http"
	"strings"

	auth "github.com/Emmrys-Jay/ecommerce-api/auth/jwt"
	"github.com/gin-gonic/gin"
)

func AuthorizeAdmin(adminName string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth_token := ctx.GetHeader("Authorization")
		if auth_token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"unauthorized": "access denied"})
			ctx.Abort()
			return
		}

		jwtToken := strings.Split(auth_token, " ")[1]
		tokenMaker, _ := auth.NewTokenMaker()

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

		if payload.Username != adminName {
			ctx.JSON(http.StatusUnauthorized, gin.H{"unauthorized": "access denied"})
			ctx.Abort()
			return
		}
	}
}
