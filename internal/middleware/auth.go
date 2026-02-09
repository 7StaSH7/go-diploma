package middleware

import (
	"net/http"
	"strings"

	"github.com/7StaSH7/practicum-diploma/internal/config"
	"github.com/7StaSH7/practicum-diploma/internal/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		userID, err := utils.ParseAccessToken(parts[1], []byte(cfg.JWTSecret))
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(UserIDKey, userID)
		c.Next()
	}
}
