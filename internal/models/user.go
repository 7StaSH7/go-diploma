package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Login        string
	PasswordHash []byte
	KDFSalt      []byte
	CreatedAt    time.Time
}
