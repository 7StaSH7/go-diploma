package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewToken(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{name: "32 bytes token", size: 32},
		{name: "16 bytes token", size: 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := NewToken(tt.size)
			require.NoError(t, err)
			require.NotEmpty(t, token)

			raw, err := base64.RawURLEncoding.DecodeString(token)
			require.NoError(t, err)
			assert.Len(t, raw, tt.size)
		})
	}
}

func TestHashToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{name: "stable hash", token: "token-value"},
		{name: "other token hash", token: "other-token"},
	}

	hashes := make(map[string][]byte, len(tests))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashToken(tt.token)
			require.Len(t, hash, sha256.Size)

			expected := sha256.Sum256([]byte(tt.token))
			assert.Equal(t, expected[:], hash)
			assert.Equal(t, hash, HashToken(tt.token))
			hashes[tt.token] = hash
		})
	}

	assert.NotEqual(t, hashes["token-value"], hashes["other-token"])
}
