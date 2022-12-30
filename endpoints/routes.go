package endpoints

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	adminMdwIndex = iota
	userMdwIndex
)

func SetupRoutes(db *mongo.Database, server *gin.Engine, mdws ...gin.HandlerFunc) {
	adminMdw := mdws[adminMdwIndex]
	userMdw := mdws[userMdwIndex]

	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Content-Length", "Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Origins"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	InitializeAdminEndpoints(db, server, adminMdw)
	InitializeCartEndpoints(db, server, userMdw)
	InitializeUserEndpoints(db, server, userMdw)
	InitializeProductEndpoints(db, server, userMdw)
	InitializeOrdersEndpoints(db, server, userMdw)

	server.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"name":    "Not Found",
			"message": "Page not found.",
			"code":    404,
			"status":  http.StatusNotFound,
		})
	})

}
