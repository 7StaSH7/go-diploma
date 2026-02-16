package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dtoauth "github.com/7StaSH7/practicum-diploma/internal/dto/auth"
	authmocks "github.com/7StaSH7/practicum-diploma/internal/handler/auth/mocks"
	authservice "github.com/7StaSH7/practicum-diploma/internal/service/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSignupSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := authmocks.NewMockService(ctrl)
	h := New(mockService)

	result := dtoauth.AuthResult{
		UserID:       uuid.New(),
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		KDFSalt:      []byte("salt"),
	}

	mockService.EXPECT().Signup(gomock.Any(), "user", "pass").Return(result, nil)

	r := gin.New()
	r.POST("/signup", h.Signup)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(`{"login":"user","password":"pass"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp dtoauth.AuthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, result.UserID.String(), resp.UserID)
	assert.Equal(t, result.AccessToken, resp.AccessToken)
	assert.Equal(t, result.RefreshToken, resp.RefreshToken)
	assert.Equal(t, base64.StdEncoding.EncodeToString(result.KDFSalt), resp.KDFSalt)
}

func TestSignupBadRequestOnInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := New(authmocks.NewMockService(ctrl))

	r := gin.New()
	r.POST("/signup", h.Signup)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(`{"login":`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSignupInternalServerErrorOnServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := authmocks.NewMockService(ctrl)
	h := New(mockService)

	mockService.EXPECT().Signup(gomock.Any(), "user", "pass").Return(dtoauth.AuthResult{}, errors.New("db error"))

	r := gin.New()
	r.POST("/signup", h.Signup)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(`{"login":"user","password":"pass"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSigninUnauthorizedOnInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := authmocks.NewMockService(ctrl)
	h := New(mockService)

	mockService.EXPECT().Signin(gomock.Any(), "user", "pass").Return(dtoauth.AuthResult{}, authservice.ErrInvalidCredentials)

	r := gin.New()
	r.POST("/signin", h.Signin)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/signin", strings.NewReader(`{"login":"user","password":"pass"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSigninInternalServerErrorOnServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := authmocks.NewMockService(ctrl)
	h := New(mockService)

	mockService.EXPECT().Signin(gomock.Any(), "user", "pass").Return(dtoauth.AuthResult{}, errors.New("db error"))

	r := gin.New()
	r.POST("/signin", h.Signin)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/signin", strings.NewReader(`{"login":"user","password":"pass"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRefreshUnauthorizedOnInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := authmocks.NewMockService(ctrl)
	h := New(mockService)

	mockService.EXPECT().Refresh(gomock.Any(), "refresh-token").Return(dtoauth.AuthResult{}, authservice.ErrInvalidCredentials)

	r := gin.New()
	r.POST("/refresh", h.Refresh)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(`{"refresh_token":"refresh-token"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefreshInternalServerErrorOnServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := authmocks.NewMockService(ctrl)
	h := New(mockService)

	mockService.EXPECT().Refresh(gomock.Any(), "refresh-token").Return(dtoauth.AuthResult{}, errors.New("db error"))

	r := gin.New()
	r.POST("/refresh", h.Refresh)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(`{"refresh_token":"refresh-token"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
