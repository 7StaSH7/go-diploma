package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"

	"golang.org/x/crypto/argon2"
)

func HashPassword(password string) ([]byte, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	encoded := make([]byte, 0, argonSaltLen+argonKeyLen)
	encoded = append(encoded, salt...)
	encoded = append(encoded, hash...)
	return encoded, nil
}

func VerifyPassword(password string, encoded []byte) (bool, error) {
	if len(encoded) != argonSaltLen+argonKeyLen {
		return false, errors.New("invalid password hash")
	}
	salt := encoded[:argonSaltLen]
	stored := encoded[argonSaltLen:]
	computed := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	if subtle.ConstantTimeCompare(stored, computed) == 1 {
		return true, nil
	}
	return false, nil
}

func NewSalt(size int) ([]byte, error) {
	salt := make([]byte, size)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}
