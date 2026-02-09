package models

import (
	"time"

	"github.com/google/uuid"
)

type MetaOpen struct {
	Title string   `json:"title"`
	Tags  []string `json:"tags,omitempty"`
	Site  string   `json:"site,omitempty"`
}

type Secret struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Type       string
	MetaOpen   MetaOpen
	Ciphertext []byte
	Version    int64
	UpdatedAt  time.Time
}
