package util

import "github.com/gin-gonic/gin"

// errorResponse returns verbose error response
func ErrorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
