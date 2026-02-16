package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/7StaSH7/practicum-diploma/internal/models"
	"github.com/google/uuid"
)

type TokenRepository interface {
	Create(ctx context.Context, token models.RefreshToken) error
	GetByHash(ctx context.Context, hash []byte) (models.RefreshToken, error)
	Rotate(ctx context.Context, hash []byte, now time.Time, next models.RefreshToken) (models.RefreshToken, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type tokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(ctx context.Context, token models.RefreshToken) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	)
	return err
}

func (r *tokenRepository) GetByHash(ctx context.Context, hash []byte) (models.RefreshToken, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at FROM refresh_tokens WHERE token_hash = $1`,
		hash,
	)
	var token models.RefreshToken
	if err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt); err != nil {
		return models.RefreshToken{}, err
	}
	return token, nil
}

func (r *tokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM refresh_tokens WHERE id = $1`,
		id,
	)
	return err
}

func (r *tokenRepository) Rotate(ctx context.Context, hash []byte, now time.Time, next models.RefreshToken) (models.RefreshToken, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return models.RefreshToken{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	row := tx.QueryRowContext(
		ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at
		 FROM refresh_tokens
		 WHERE token_hash = $1
		 FOR UPDATE`,
		hash,
	)
	var current models.RefreshToken
	if err := row.Scan(&current.ID, &current.UserID, &current.TokenHash, &current.ExpiresAt, &current.CreatedAt); err != nil {
		return models.RefreshToken{}, err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE id = $1`, current.ID)
	if err != nil {
		return models.RefreshToken{}, err
	}
	if !current.ExpiresAt.After(now) {
		if err := tx.Commit(); err != nil {
			return models.RefreshToken{}, err
		}
		committed = true
		return models.RefreshToken{}, sql.ErrNoRows
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		next.ID,
		current.UserID,
		next.TokenHash,
		next.ExpiresAt,
		next.CreatedAt,
	)
	if err != nil {
		return models.RefreshToken{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.RefreshToken{}, err
	}
	committed = true
	return current, nil
}
