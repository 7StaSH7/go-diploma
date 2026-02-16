package auth

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/7StaSH7/practicum-diploma/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewTokenRepository(db)
	token := models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: []byte("hash"),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`)).
		WithArgs(token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), token)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepositoryGetByHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewTokenRepository(db)
	token := models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: []byte("hash"),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, token_hash, expires_at, created_at FROM refresh_tokens WHERE token_hash = $1`)).
		WithArgs(token.TokenHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at"}).
			AddRow(token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt))

	got, err := repo.GetByHash(context.Background(), token.TokenHash)
	require.NoError(t, err)
	assert.Equal(t, token, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepositoryRotateSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewTokenRepository(db)
	now := time.Now().UTC()
	hash := []byte("current-hash")
	current := models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: hash,
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now.Add(-time.Hour),
	}
	next := models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: []byte("next-hash"),
		ExpiresAt: now.Add(2 * time.Hour),
		CreatedAt: now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, token_hash, expires_at, created_at
		 FROM refresh_tokens
		 WHERE token_hash = $1
		 FOR UPDATE`)).
		WithArgs(hash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at"}).
			AddRow(current.ID, current.UserID, current.TokenHash, current.ExpiresAt, current.CreatedAt))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM refresh_tokens WHERE id = $1`)).
		WithArgs(current.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`)).
		WithArgs(next.ID, current.UserID, next.TokenHash, next.ExpiresAt, next.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	got, err := repo.Rotate(context.Background(), hash, now, next)
	require.NoError(t, err)
	assert.Equal(t, current, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepositoryRotateExpiredToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewTokenRepository(db)
	now := time.Now().UTC()
	hash := []byte("current-hash")
	current := models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: hash,
		ExpiresAt: now.Add(-time.Minute),
		CreatedAt: now.Add(-time.Hour),
	}
	next := models.RefreshToken{ID: uuid.New(), TokenHash: []byte("next")}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, token_hash, expires_at, created_at
		 FROM refresh_tokens
		 WHERE token_hash = $1
		 FOR UPDATE`)).
		WithArgs(hash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at"}).
			AddRow(current.ID, current.UserID, current.TokenHash, current.ExpiresAt, current.CreatedAt))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM refresh_tokens WHERE id = $1`)).
		WithArgs(current.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	_, err = repo.Rotate(context.Background(), hash, now, next)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepositoryRotateInsertErrorRollsBack(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewTokenRepository(db)
	now := time.Now().UTC()
	hash := []byte("current-hash")
	current := models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: hash,
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now.Add(-time.Hour),
	}
	next := models.RefreshToken{
		ID:        uuid.New(),
		TokenHash: []byte("next-hash"),
		ExpiresAt: now.Add(2 * time.Hour),
		CreatedAt: now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, token_hash, expires_at, created_at
		 FROM refresh_tokens
		 WHERE token_hash = $1
		 FOR UPDATE`)).
		WithArgs(hash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at"}).
			AddRow(current.ID, current.UserID, current.TokenHash, current.ExpiresAt, current.CreatedAt))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM refresh_tokens WHERE id = $1`)).
		WithArgs(current.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`)).
		WithArgs(next.ID, current.UserID, next.TokenHash, next.ExpiresAt, next.CreatedAt).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	_, err = repo.Rotate(context.Background(), hash, now, next)
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	require.NoError(t, mock.ExpectationsWereMet())
}
