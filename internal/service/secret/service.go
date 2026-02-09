package secret

//go:generate go run go.uber.org/mock/mockgen@latest -destination=./mocks/secret_repository_mock.go -package=mocks github.com/7StaSH7/practicum-diploma/internal/repository/secret SecretRepository

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	dtosecret "github.com/7StaSH7/practicum-diploma/internal/dto/secret"
	"github.com/7StaSH7/practicum-diploma/internal/models"
	secretrepository "github.com/7StaSH7/practicum-diploma/internal/repository/secret"
	"github.com/google/uuid"
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrNotFound          = errors.New("secret not found")
)

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, payload dtosecret.SecretInput) (models.Secret, error)
	Update(ctx context.Context, userID uuid.UUID, secretID uuid.UUID, payload dtosecret.SecretInput) (models.Secret, error)
	Delete(ctx context.Context, userID uuid.UUID, secretID uuid.UUID) error
	Get(ctx context.Context, userID uuid.UUID, secretID uuid.UUID) (models.Secret, error)
	ListSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]models.Secret, error)
}

type service struct {
	secrets secretrepository.SecretRepository
}

func NewService(secrets secretrepository.SecretRepository) Service {
	return &service{
		secrets: secrets,
	}
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, payload dtosecret.SecretInput) (models.Secret, error) {
	data, err := decodeCiphertext(payload.Ciphertext)
	if err != nil {
		return models.Secret{}, ErrInvalidCiphertext
	}
	secret := models.Secret{
		ID:         uuid.New(),
		UserID:     userID,
		Type:       payload.Type,
		MetaOpen:   payload.MetaOpen,
		Ciphertext: data,
		Version:    1,
		UpdatedAt:  time.Now().UTC(),
	}
	if err := s.secrets.Create(ctx, secret); err != nil {
		return models.Secret{}, err
	}
	return secret, nil
}

func (s *service) Update(ctx context.Context, userID uuid.UUID, secretID uuid.UUID, payload dtosecret.SecretInput) (models.Secret, error) {
	data, err := decodeCiphertext(payload.Ciphertext)
	if err != nil {
		return models.Secret{}, ErrInvalidCiphertext
	}
	current, err := s.secrets.Get(ctx, secretID, userID)
	if err != nil {
		return models.Secret{}, mapNotFound(err)
	}
	current.Type = payload.Type
	current.MetaOpen = payload.MetaOpen
	current.Ciphertext = data
	current.Version++
	current.UpdatedAt = time.Now().UTC()
	if err := s.secrets.Update(ctx, current); err != nil {
		return models.Secret{}, mapNotFound(err)
	}
	return current, nil
}

func (s *service) Delete(ctx context.Context, userID uuid.UUID, secretID uuid.UUID) error {
	return mapNotFound(s.secrets.Delete(ctx, secretID, userID))
}

func (s *service) Get(ctx context.Context, userID uuid.UUID, secretID uuid.UUID) (models.Secret, error) {
	secret, err := s.secrets.Get(ctx, secretID, userID)
	if err != nil {
		return models.Secret{}, mapNotFound(err)
	}
	return secret, nil
}

func (s *service) ListSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]models.Secret, error) {
	return s.secrets.ListSince(ctx, userID, since)
}

func decodeCiphertext(ciphertext string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(ciphertext)
}

func mapNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
