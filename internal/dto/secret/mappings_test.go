package secret

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/7StaSH7/practicum-diploma/internal/models"
	"github.com/google/uuid"
)

func TestToSecretResponseEncodesCiphertext(t *testing.T) {
	secret := models.Secret{
		ID:         uuid.New(),
		Type:       "note",
		MetaOpen:   models.MetaOpen{Title: "hello"},
		Ciphertext: []byte("payload"),
		Version:    3,
		UpdatedAt:  time.Date(2026, time.February, 6, 9, 0, 0, 0, time.UTC),
	}

	resp := ToSecretResponse(secret)
	if resp.UpdatedAt != "2026-02-06T09:00:00Z" {
		t.Fatalf("unexpected updated_at: %s", resp.UpdatedAt)
	}
	if resp.Ciphertext != base64.StdEncoding.EncodeToString([]byte("payload")) {
		t.Fatalf("unexpected ciphertext: %s", resp.Ciphertext)
	}
}
