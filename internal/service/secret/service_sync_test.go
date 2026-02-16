package secret

import (
	"context"
	"database/sql"
	"encoding/base64"
	"sort"
	"testing"
	"time"

	dtosecret "github.com/7StaSH7/practicum-diploma/internal/dto/secret"
	"github.com/7StaSH7/practicum-diploma/internal/models"
	"github.com/google/uuid"
)

type memorySecretRepo struct {
	items map[uuid.UUID]models.Secret
}

func newMemorySecretRepo() *memorySecretRepo {
	return &memorySecretRepo{
		items: make(map[uuid.UUID]models.Secret),
	}
}

func (m *memorySecretRepo) Create(_ context.Context, secret models.Secret) error {
	m.items[secret.ID] = secret
	return nil
}

func (m *memorySecretRepo) Update(_ context.Context, secret models.Secret) error {
	current, ok := m.items[secret.ID]
	if !ok || current.UserID != secret.UserID {
		return sql.ErrNoRows
	}
	m.items[secret.ID] = secret
	return nil
}

func (m *memorySecretRepo) Get(_ context.Context, id uuid.UUID, userID uuid.UUID) (models.Secret, error) {
	secret, ok := m.items[id]
	if !ok || secret.UserID != userID {
		return models.Secret{}, sql.ErrNoRows
	}
	return secret, nil
}

func (m *memorySecretRepo) ListSince(_ context.Context, userID uuid.UUID, since time.Time) ([]models.Secret, error) {
	out := make([]models.Secret, 0)
	for _, secret := range m.items {
		if secret.UserID != userID {
			continue
		}
		if secret.UpdatedAt.After(since) {
			out = append(out, secret)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.Before(out[j].UpdatedAt)
	})
	return out, nil
}

func (m *memorySecretRepo) Delete(_ context.Context, id uuid.UUID, userID uuid.UUID) error {
	secret, ok := m.items[id]
	if !ok || secret.UserID != userID {
		return sql.ErrNoRows
	}
	delete(m.items, id)
	return nil
}

func TestDeleteRemovesSecretCompletely(t *testing.T) {
	repo := newMemorySecretRepo()
	service := NewService(repo)
	userID := uuid.New()

	payload := dtosecret.SecretInput{
		Type:       "note",
		Ciphertext: base64.StdEncoding.EncodeToString([]byte("payload")),
		MetaOpen: models.MetaOpen{
			Title: "record",
		},
	}

	created, err := service.Create(context.Background(), userID, payload)
	if err != nil {
		t.Fatalf("create secret: %v", err)
	}

	checkpoint := time.Now().UTC()
	time.Sleep(2 * time.Millisecond)
	if err := service.Delete(context.Background(), userID, created.ID); err != nil {
		t.Fatalf("delete secret: %v", err)
	}

	if _, err := service.Get(context.Background(), userID, created.ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got: %v", err)
	}

	changes, err := service.ListSince(context.Background(), userID, checkpoint)
	if err != nil {
		t.Fatalf("list since: %v", err)
	}
	if len(changes) != 0 {
		t.Fatalf("expected no changes after hard delete, got %d", len(changes))
	}
}
