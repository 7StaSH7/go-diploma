package secret

import (
	"context"
	"database/sql"
	"encoding/base64"
	"testing"
	"time"

	dtosecret "github.com/7StaSH7/practicum-diploma/internal/dto/secret"
	"github.com/7StaSH7/practicum-diploma/internal/models"
	secretmocks "github.com/7StaSH7/practicum-diploma/internal/service/secret/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateReturnsInvalidCiphertext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := secretmocks.NewMockSecretRepository(ctrl)
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), uuid.New(), dtosecret.SecretInput{Ciphertext: "!!!"})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCiphertext)
}

func TestCreateStoresDecodedCiphertext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := secretmocks.NewMockSecretRepository(ctrl)
	svc := NewService(repo)
	userID := uuid.New()

	payload := dtosecret.SecretInput{
		Type:       "note",
		MetaOpen:   models.MetaOpen{Title: "title"},
		Ciphertext: base64.StdEncoding.EncodeToString([]byte("cipher")),
	}

	repo.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(models.Secret{})).DoAndReturn(
		func(_ context.Context, secret models.Secret) error {
			assert.Equal(t, userID, secret.UserID)
			assert.Equal(t, payload.Type, secret.Type)
			assert.Equal(t, payload.MetaOpen, secret.MetaOpen)
			assert.Equal(t, []byte("cipher"), secret.Ciphertext)
			assert.Equal(t, int64(1), secret.Version)
			assert.WithinDuration(t, time.Now().UTC(), secret.UpdatedAt, 2*time.Second)
			return nil
		},
	)

	created, err := svc.Create(context.Background(), userID, payload)
	require.NoError(t, err)
	assert.Equal(t, []byte("cipher"), created.Ciphertext)
	assert.Equal(t, int64(1), created.Version)
}

func TestUpdateReturnsNotFoundWhenSecretMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := secretmocks.NewMockSecretRepository(ctrl)
	svc := NewService(repo)

	repo.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(models.Secret{}, sql.ErrNoRows)

	_, err := svc.Update(context.Background(), uuid.New(), uuid.New(), dtosecret.SecretInput{Ciphertext: base64.StdEncoding.EncodeToString([]byte("cipher"))})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDeleteReturnsNotFoundWhenMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := secretmocks.NewMockSecretRepository(ctrl)
	svc := NewService(repo)

	repo.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(sql.ErrNoRows)

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestListSinceDelegatesToRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := secretmocks.NewMockSecretRepository(ctrl)
	svc := NewService(repo)
	userID := uuid.New()
	since := time.Now().UTC().Add(-time.Hour)
	expected := []models.Secret{{ID: uuid.New(), UserID: userID, Version: 2}}

	repo.EXPECT().ListSince(gomock.Any(), userID, since).Return(expected, nil)

	got, err := svc.ListSince(context.Background(), userID, since)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}
