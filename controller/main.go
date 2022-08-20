package controller

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type ServerDB struct {
	Db     *mongo.Database
	Server *gin.Engine
}
