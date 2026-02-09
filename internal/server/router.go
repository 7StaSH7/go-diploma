package server

import (
	"github.com/7StaSH7/practicum-diploma/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter(log *zap.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware(log))
	return router
}
