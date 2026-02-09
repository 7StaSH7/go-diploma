package secret

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/7StaSH7/practicum-diploma/internal/models"
	"github.com/google/uuid"
)

type SecretRepository interface {
	Create(ctx context.Context, secret models.Secret) error
	Update(ctx context.Context, secret models.Secret) error
	Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (models.Secret, error)
	ListSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]models.Secret, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type secretRepository struct {
	db *sql.DB
}

func NewSecretRepository(db *sql.DB) SecretRepository {
	return &secretRepository{db: db}
}

func (r *secretRepository) Create(ctx context.Context, secret models.Secret) error {
	metaBytes, err := json.Marshal(secret.MetaOpen)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO secrets (id, user_id, type, meta_open, ciphertext, version, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		secret.ID,
		secret.UserID,
		secret.Type,
		metaBytes,
		secret.Ciphertext,
		secret.Version,
		secret.UpdatedAt,
	)
	return err
}

func (r *secretRepository) Update(ctx context.Context, secret models.Secret) error {
	metaBytes, err := json.Marshal(secret.MetaOpen)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE secrets
		 SET type = $1, meta_open = $2, ciphertext = $3, version = $4, updated_at = $5
		 WHERE id = $6 AND user_id = $7`,
		secret.Type,
		metaBytes,
		secret.Ciphertext,
		secret.Version,
		secret.UpdatedAt,
		secret.ID,
		secret.UserID,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *secretRepository) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (models.Secret, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, type, meta_open, ciphertext, version, updated_at
		 FROM secrets WHERE id = $1 AND user_id = $2`,
		id,
		userID,
	)
	return scanSecret(row)
}

func (r *secretRepository) ListSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]models.Secret, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, user_id, type, meta_open, ciphertext, version, updated_at
		 FROM secrets WHERE user_id = $1 AND updated_at > $2
		 ORDER BY updated_at ASC`,
		userID,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []models.Secret
	for rows.Next() {
		secret, err := scanSecret(rows)
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, secret)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return secrets, nil
}

func (r *secretRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM secrets WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanSecret(row scanner) (models.Secret, error) {
	var secret models.Secret
	var metaBytes []byte
	if err := row.Scan(
		&secret.ID,
		&secret.UserID,
		&secret.Type,
		&metaBytes,
		&secret.Ciphertext,
		&secret.Version,
		&secret.UpdatedAt,
	); err != nil {
		return models.Secret{}, err
	}
	if len(metaBytes) > 0 {
		if err := json.Unmarshal(metaBytes, &secret.MetaOpen); err != nil {
			return models.Secret{}, err
		}
	}
	return secret, nil
}
