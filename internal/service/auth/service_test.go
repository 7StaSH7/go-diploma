package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/7StaSH7/practicum-diploma/internal/config"
	"github.com/7StaSH7/practicum-diploma/internal/models"
	authmocks "github.com/7StaSH7/practicum-diploma/internal/service/auth/mocks"
	"github.com/7StaSH7/practicum-diploma/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestSigninReturnsInvalidCredentialsWhenUserMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := authmocks.NewMockUserRepository(ctrl)
	tokens := authmocks.NewMockTokenRepository(ctrl)
	svc := NewService(users, tokens, testConfig(), zap.NewNop())

	users.EXPECT().GetByLogin(gomock.Any(), "missing").Return(models.User{}, sql.ErrNoRows)

	_, err := svc.Signin(context.Background(), "missing", "password")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestSigninReturnsInvalidCredentialsWhenPasswordMismatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hash, err := utils.HashPassword("correct-password")
	require.NoError(t, err)

	users := authmocks.NewMockUserRepository(ctrl)
	tokens := authmocks.NewMockTokenRepository(ctrl)
	svc := NewService(users, tokens, testConfig(), zap.NewNop())

	users.EXPECT().GetByLogin(gomock.Any(), "user").Return(models.User{
		ID:           uuid.New(),
		Login:        "user",
		PasswordHash: hash,
		KDFSalt:      []byte("salt"),
	}, nil)

	_, err = svc.Signin(context.Background(), "user", "wrong-password")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestSignupCreatesUserAndIssuesTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := authmocks.NewMockUserRepository(ctrl)
	tokens := authmocks.NewMockTokenRepository(ctrl)
	svc := NewService(users, tokens, testConfig(), zap.NewNop())

	var createdUser models.User
	users.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(models.User{})).DoAndReturn(
		func(_ context.Context, user models.User) error {
			createdUser = user
			return nil
		},
	)
	tokens.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(models.RefreshToken{})).DoAndReturn(
		func(_ context.Context, token models.RefreshToken) error {
			assert.Equal(t, createdUser.ID, token.UserID)
			assert.NotEmpty(t, token.TokenHash)
			return nil
		},
	)

	result, err := svc.Signup(context.Background(), "user", "password")
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, createdUser.ID)
	assert.Equal(t, "user", createdUser.Login)
	assert.NotEmpty(t, createdUser.PasswordHash)
	assert.NotEmpty(t, createdUser.KDFSalt)
	assert.Equal(t, createdUser.ID, result.UserID)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, createdUser.KDFSalt, result.KDFSalt)
}

func TestRefreshReturnsInvalidCredentialsWhenRotateMisses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := authmocks.NewMockUserRepository(ctrl)
	tokens := authmocks.NewMockTokenRepository(ctrl)
	svc := NewService(users, tokens, testConfig(), zap.NewNop())

	tokens.EXPECT().Rotate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.AssignableToTypeOf(models.RefreshToken{})).Return(models.RefreshToken{}, sql.ErrNoRows)

	_, err := svc.Refresh(context.Background(), "refresh-token")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestRefreshReturnsUserDataAndNewTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	users := authmocks.NewMockUserRepository(ctrl)
	tokens := authmocks.NewMockTokenRepository(ctrl)
	svc := NewService(users, tokens, testConfig(), zap.NewNop())

	userID := uuid.New()
	rotated := models.RefreshToken{ID: uuid.New(), UserID: userID}
	tokens.EXPECT().Rotate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.AssignableToTypeOf(models.RefreshToken{})).Return(rotated, nil)
	users.EXPECT().GetByID(gomock.Any(), userID).Return(models.User{ID: userID, KDFSalt: []byte("salt")}, nil)

	result, err := svc.Refresh(context.Background(), "refresh-token")
	require.NoError(t, err)
	assert.Equal(t, userID, result.UserID)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, []byte("salt"), result.KDFSalt)
}

func testConfig() config.Config {
	return config.Config{
		JWTSecret:  "test-secret",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * time.Hour,
	}
}
