package secret

//go:generate go run go.uber.org/mock/mockgen@latest -destination=./mocks/secret_service_mock.go -package=mocks github.com/7StaSH7/practicum-diploma/internal/service/secret Service

import (
	"errors"
	"net/http"
	"time"

	dtosecret "github.com/7StaSH7/practicum-diploma/internal/dto/secret"
	"github.com/7StaSH7/practicum-diploma/internal/middleware"
	"github.com/7StaSH7/practicum-diploma/internal/models"
	secretservice "github.com/7StaSH7/practicum-diploma/internal/service/secret"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler interface {
	CreateSecret(c *gin.Context)
	UpdateSecret(c *gin.Context)
	DeleteSecret(c *gin.Context)
	GetSecret(c *gin.Context)
	ListSecrets(c *gin.Context)
}

type handler struct {
	service secretservice.Service
}

func New(service secretservice.Service) Handler {
	return &handler{
		service: service,
	}
}

func (h *handler) CreateSecret(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var payload dtosecret.SecretPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		_ = c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	created, err := h.service.Create(c.Request.Context(), userID, dtosecret.ToSecretInput(payload))
	if err != nil {
		_ = c.Error(err)
		if errors.Is(err, secretservice.ErrInvalidCiphertext) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	respondSecret(c, created)
}

func (h *handler) UpdateSecret(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	secretID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var payload dtosecret.SecretPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		_ = c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	current, err := h.service.Update(c.Request.Context(), userID, secretID, dtosecret.ToSecretInput(payload))
	if err != nil {
		_ = c.Error(err)
		if errors.Is(err, secretservice.ErrInvalidCiphertext) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if errors.Is(err, secretservice.ErrNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	respondSecret(c, current)
}

func (h *handler) DeleteSecret(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	secretID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if err := h.service.Delete(c.Request.Context(), userID, secretID); err != nil {
		_ = c.Error(err)
		if errors.Is(err, secretservice.ErrNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *handler) GetSecret(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	secretID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	found, err := h.service.Get(c.Request.Context(), userID, secretID)
	if err != nil {
		_ = c.Error(err)
		if errors.Is(err, secretservice.ErrNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	respondSecret(c, found)
}

func (h *handler) ListSecrets(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	since := time.Time{}
	if sinceRaw := c.Query("since"); sinceRaw != "" {
		parsed, err := time.Parse(time.RFC3339, sinceRaw)
		if err != nil {
			_ = c.Error(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		since = parsed
	}
	secrets, err := h.service.ListSince(c.Request.Context(), userID, since)
	if err != nil {
		_ = c.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	responses := make([]dtosecret.SecretResponse, 0, len(secrets))
	for _, secret := range secrets {
		responses = append(responses, dtosecret.ToSecretResponse(secret))
	}
	c.JSON(http.StatusOK, responses)
}

func respondSecret(c *gin.Context, secret models.Secret) {
	c.JSON(http.StatusOK, dtosecret.ToSecretResponse(secret))
}

func userIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	value, ok := c.Get(middleware.UserIDKey)
	if !ok {
		return uuid.UUID{}, false
	}
	parsed, err := uuid.Parse(value.(string))
	if err != nil {
		return uuid.UUID{}, false
	}
	return parsed, true
}
