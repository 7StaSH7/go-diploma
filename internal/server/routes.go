package server

import (
	"github.com/7StaSH7/practicum-diploma/internal/config"
	handlerauth "github.com/7StaSH7/practicum-diploma/internal/handler/auth"
	handlersecret "github.com/7StaSH7/practicum-diploma/internal/handler/secret"
	"github.com/7StaSH7/practicum-diploma/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, cfg config.Config, authHandlers handlerauth.Handler, secretHandlers handlersecret.Handler) {
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/signup", authHandlers.Signup)
		authRoutes.POST("/signin", authHandlers.Signin)
		authRoutes.POST("/refresh", authHandlers.Refresh)
	}

	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		protected.GET("/secrets", secretHandlers.ListSecrets)
		protected.POST("/secrets", secretHandlers.CreateSecret)
		protected.GET("/secrets/:id", secretHandlers.GetSecret)
		protected.PUT("/secrets/:id", secretHandlers.UpdateSecret)
		protected.DELETE("/secrets/:id", secretHandlers.DeleteSecret)
	}
}
