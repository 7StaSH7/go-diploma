package secret

import (
	"encoding/base64"
	"time"

	"github.com/7StaSH7/practicum-diploma/internal/models"
)

func ToSecretResponse(secret models.Secret) SecretResponse {
	updated := secret.UpdatedAt.UTC().Format(time.RFC3339)
	return SecretResponse{
		ID:         secret.ID.String(),
		Type:       secret.Type,
		MetaOpen:   secret.MetaOpen,
		Ciphertext: base64.StdEncoding.EncodeToString(secret.Ciphertext),
		Version:    secret.Version,
		UpdatedAt:  updated,
	}
}

func ToSecretInput(payload SecretPayload) SecretInput {
	return SecretInput{
		Type:       payload.Type,
		MetaOpen:   payload.MetaOpen,
		Ciphertext: payload.Ciphertext,
	}
}
