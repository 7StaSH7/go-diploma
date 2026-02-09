package secret

import "github.com/7StaSH7/practicum-diploma/internal/models"

type SecretPayload struct {
	Type       string          `json:"type"`
	MetaOpen   models.MetaOpen `json:"meta_open"`
	Ciphertext string          `json:"ciphertext"`
}

type SecretInput struct {
	Type       string
	MetaOpen   models.MetaOpen
	Ciphertext string
}

type SecretResponse struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	MetaOpen   models.MetaOpen `json:"meta_open"`
	Ciphertext string          `json:"ciphertext"`
	Version    int64           `json:"version"`
	UpdatedAt  string          `json:"updated_at"`
}
