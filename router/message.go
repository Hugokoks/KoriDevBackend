package router

import (
	"koridev/handlers"

	"github.com/gin-gonic/gin"
)

func Message(g *gin.RouterGroup) {

	g.POST("/message", handlers.PostMessage)

}
