package secret

import (
	"context"
	"database/sql"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/7StaSH7/practicum-diploma/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSecretRepository(db)
	secret := models.Secret{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		Type:       "note",
		MetaOpen:   models.MetaOpen{Title: "title", Tags: []string{"a"}},
		Ciphertext: []byte("cipher"),
		Version:    1,
		UpdatedAt:  time.Now().UTC(),
	}

	metaBytes, err := json.Marshal(secret.MetaOpen)
	require.NoError(t, err)

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO secrets (id, user_id, type, meta_open, ciphertext, version, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`)).
		WithArgs(secret.ID, secret.UserID, secret.Type, metaBytes, secret.Ciphertext, secret.Version, secret.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), secret)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepositoryUpdateNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSecretRepository(db)
	secret := models.Secret{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		Type:       "note",
		MetaOpen:   models.MetaOpen{Title: "title"},
		Ciphertext: []byte("cipher"),
		Version:    2,
		UpdatedAt:  time.Now().UTC(),
	}

	metaBytes, err := json.Marshal(secret.MetaOpen)
	require.NoError(t, err)

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE secrets
		 SET type = $1, meta_open = $2, ciphertext = $3, version = $4, updated_at = $5
		 WHERE id = $6 AND user_id = $7`)).
		WithArgs(secret.Type, metaBytes, secret.Ciphertext, secret.Version, secret.UpdatedAt, secret.ID, secret.UserID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(context.Background(), secret)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepositoryGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSecretRepository(db)
	secretID := uuid.New()
	userID := uuid.New()
	updatedAt := time.Now().UTC().Truncate(time.Second)
	metaBytes := []byte(`{"title":"title","tags":["x"]}`)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, type, meta_open, ciphertext, version, updated_at
		 FROM secrets WHERE id = $1 AND user_id = $2`)).
		WithArgs(secretID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "type", "meta_open", "ciphertext", "version", "updated_at"}).
			AddRow(secretID, userID, "note", metaBytes, []byte("cipher"), int64(3), updatedAt))

	got, err := repo.Get(context.Background(), secretID, userID)
	require.NoError(t, err)
	assert.Equal(t, secretID, got.ID)
	assert.Equal(t, userID, got.UserID)
	assert.Equal(t, "note", got.Type)
	assert.Equal(t, []byte("cipher"), got.Ciphertext)
	assert.Equal(t, int64(3), got.Version)
	assert.Equal(t, updatedAt, got.UpdatedAt)
	assert.Equal(t, models.MetaOpen{Title: "title", Tags: []string{"x"}}, got.MetaOpen)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepositoryListSince(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSecretRepository(db)
	userID := uuid.New()
	since := time.Now().UTC().Add(-time.Hour)
	updatedAt := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, type, meta_open, ciphertext, version, updated_at
		 FROM secrets WHERE user_id = $1 AND updated_at > $2
		 ORDER BY updated_at ASC`)).
		WithArgs(userID, since).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "type", "meta_open", "ciphertext", "version", "updated_at"}).
			AddRow(uuid.New(), userID, "note", []byte(`{"title":"one"}`), []byte("a"), int64(1), updatedAt.Add(-time.Minute)).
			AddRow(uuid.New(), userID, "card", []byte(`{"title":"two"}`), []byte("b"), int64(2), updatedAt))

	items, err := repo.ListSince(context.Background(), userID, since)
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "one", items[0].MetaOpen.Title)
	assert.Equal(t, "two", items[1].MetaOpen.Title)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSecretRepositoryDeleteNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewSecretRepository(db)
	secretID := uuid.New()
	userID := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM secrets WHERE id = $1 AND user_id = $2`)).
		WithArgs(secretID, userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(context.Background(), secretID, userID)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}
