package api

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
)

type Server struct {
	DB     *mgo.Database
	Server *gin.Engine
}
