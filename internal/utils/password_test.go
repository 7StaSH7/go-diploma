package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyPassword(t *testing.T) {
	validHash, err := HashPassword("strong-password")
	require.NoError(t, err)

	tests := []struct {
		name          string
		password      string
		hash          []byte
		wantOK        bool
		wantErr       bool
		checkHashSize bool
	}{
		{name: "correct password", password: "strong-password", hash: validHash, wantOK: true, checkHashSize: true},
		{name: "wrong password", password: "wrong-password", hash: validHash, wantOK: false},
		{name: "invalid hash", password: "password", hash: []byte("short"), wantOK: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.checkHashSize {
				require.Len(t, tt.hash, argonSaltLen+argonKeyLen)
			}

			gotOK, err := VerifyPassword(tt.password, tt.hash)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantOK, gotOK)
		})
	}
}

func TestNewSalt(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{name: "16 bytes", size: 16},
		{name: "32 bytes", size: 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saltA, err := NewSalt(tt.size)
			require.NoError(t, err)
			require.Len(t, saltA, tt.size)

			saltB, err := NewSalt(tt.size)
			require.NoError(t, err)
			require.Len(t, saltB, tt.size)

			assert.NotEqual(t, saltA, saltB)
		})
	}
}
