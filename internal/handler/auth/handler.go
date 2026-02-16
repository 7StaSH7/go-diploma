package auth

//go:generate go run go.uber.org/mock/mockgen@latest -destination=./mocks/auth_service_mock.go -package=mocks github.com/7StaSH7/practicum-diploma/internal/service/auth Service

import (
	"errors"
	"net/http"

	dtoauth "github.com/7StaSH7/practicum-diploma/internal/dto/auth"
	authservice "github.com/7StaSH7/practicum-diploma/internal/service/auth"
	"github.com/gin-gonic/gin"
)

type Handler interface {
	Signup(c *gin.Context)
	Signin(c *gin.Context)
	Refresh(c *gin.Context)
}

type handler struct {
	service authservice.Service
}

func New(service authservice.Service) Handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Signup(c *gin.Context) {
	var req dtoauth.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	result, err := h.service.Signup(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		_ = c.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, dtoauth.ToAuthResponse(result))
}

func (h *handler) Signin(c *gin.Context) {
	var req dtoauth.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	result, err := h.service.Signin(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		_ = c.Error(err)
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, dtoauth.ToAuthResponse(result))
}

func (h *handler) Refresh(c *gin.Context) {
	var req dtoauth.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	result, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		_ = c.Error(err)
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, dtoauth.ToAuthResponse(result))
}
