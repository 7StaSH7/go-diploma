package secret

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dtosecret "github.com/7StaSH7/practicum-diploma/internal/dto/secret"
	secretmocks "github.com/7StaSH7/practicum-diploma/internal/handler/secret/mocks"
	"github.com/7StaSH7/practicum-diploma/internal/middleware"
	"github.com/7StaSH7/practicum-diploma/internal/models"
	secretservice "github.com/7StaSH7/practicum-diploma/internal/service/secret"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateSecretUnauthorizedWithoutUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := New(secretmocks.NewMockService(ctrl))
	r := gin.New()
	r.POST("/secrets", h.CreateSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/secrets", strings.NewReader(`{"type":"note","ciphertext":"YQ=="}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateSecretBadRequestOnInvalidCiphertext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	mockService := secretmocks.NewMockService(ctrl)
	h := New(mockService)

	mockService.EXPECT().Create(gomock.Any(), userID, gomock.Any()).Return(models.Secret{}, secretservice.ErrInvalidCiphertext)

	r := gin.New()
	r.Use(withUserID(userID))
	r.POST("/secrets", h.CreateSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/secrets", strings.NewReader(`{"type":"note","ciphertext":"!!!"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSecretNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	secretID := uuid.New()
	mockService := secretmocks.NewMockService(ctrl)
	h := New(mockService)

	mockService.EXPECT().Get(gomock.Any(), userID, secretID).Return(models.Secret{}, secretservice.ErrNotFound)

	r := gin.New()
	r.Use(withUserID(userID))
	r.GET("/secrets/:id", h.GetSecret)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secrets/"+secretID.String(), nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListSecretsBadRequestOnInvalidSince(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := New(secretmocks.NewMockService(ctrl))
	r := gin.New()
	r.Use(withUserID(uuid.New()))
	r.GET("/secrets", h.ListSecrets)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secrets?since=invalid-time", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListSecretsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	since := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	updatedAt := time.Now().UTC().Truncate(time.Second)
	secretID := uuid.New()
	mockService := secretmocks.NewMockService(ctrl)
	h := New(mockService)

	mockService.EXPECT().ListSince(gomock.Any(), userID, since).Return([]models.Secret{
		{
			ID:         secretID,
			UserID:     userID,
			Type:       "note",
			Ciphertext: []byte("secret-data"),
			Version:    2,
			UpdatedAt:  updatedAt,
		},
	}, nil)

	r := gin.New()
	r.Use(withUserID(userID))
	r.GET("/secrets", h.ListSecrets)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secrets?since="+since.Format(time.RFC3339), nil)

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response []dtosecret.SecretResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	require.Len(t, response, 1)
	assert.Equal(t, secretID.String(), response[0].ID)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("secret-data")), response[0].Ciphertext)
	assert.Equal(t, int64(2), response[0].Version)
	assert.Equal(t, updatedAt.Format(time.RFC3339), response[0].UpdatedAt)
}

func withUserID(userID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID.String())
		c.Next()
	}
}
