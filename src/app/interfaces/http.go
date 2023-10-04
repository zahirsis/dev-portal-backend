package interfaces

import "github.com/gin-gonic/gin"

type Router interface {
	gin.IRouter
}

type HttpServerContext = *gin.Context
