package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAndParseAccessTokenSuccess(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		secret []byte
		ttl    time.Duration
	}{
		{name: "valid token", userID: "user-123", secret: []byte("jwt-secret"), ttl: time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := NewAccessToken(tt.userID, tt.secret, tt.ttl)
			require.NoError(t, err)

			gotUserID, err := ParseAccessToken(token, tt.secret)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, gotUserID)
		})
	}
}

func TestParseAccessTokenErrors(t *testing.T) {
	buildNoSubjectToken := func() string {
		secret := []byte("jwt-secret")
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		}).SignedString(secret)
		require.NoError(t, err)
		return token
	}

	buildWrongSecretToken := func() string {
		token, err := NewAccessToken("user-123", []byte("secret-a"), time.Minute)
		require.NoError(t, err)
		return token
	}

	tests := []struct {
		name   string
		token  string
		secret []byte
	}{
		{name: "invalid secret", token: buildWrongSecretToken(), secret: []byte("secret-b")},
		{name: "malformed token", token: "not-a-jwt", secret: []byte("secret")},
		{name: "without subject", token: buildNoSubjectToken(), secret: []byte("jwt-secret")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseAccessToken(tt.token, tt.secret)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidToken)
		})
	}
}
