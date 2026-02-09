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

func TestUserRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(db)
	user := models.User{
		ID:           uuid.New(),
		Login:        "alice",
		PasswordHash: []byte("hash"),
		KDFSalt:      []byte("salt"),
		CreatedAt:    time.Now().UTC(),
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (id, login, password_hash, kdf_salt, created_at)
		 VALUES ($1, $2, $3, $4, $5)`)).
		WithArgs(user.ID, user.Login, user.PasswordHash, user.KDFSalt, user.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), user)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepositoryGetByLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(db)
	user := models.User{
		ID:           uuid.New(),
		Login:        "alice",
		PasswordHash: []byte("hash"),
		KDFSalt:      []byte("salt"),
		CreatedAt:    time.Now().UTC(),
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, login, password_hash, kdf_salt, created_at FROM users WHERE login = $1`)).
		WithArgs(user.Login).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password_hash", "kdf_salt", "created_at"}).
			AddRow(user.ID, user.Login, user.PasswordHash, user.KDFSalt, user.CreatedAt))

	got, err := repo.GetByLogin(context.Background(), user.Login)
	require.NoError(t, err)
	assert.Equal(t, user, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepositoryGetByIDNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := NewUserRepository(db)
	userID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, login, password_hash, kdf_salt, created_at FROM users WHERE id = $1`)).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(context.Background(), userID)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}
