package auth

import (
	"context"
	"database/sql"

	"github.com/7StaSH7/practicum-diploma/internal/models"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user models.User) error
	GetByLogin(ctx context.Context, login string) (models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user models.User) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (id, login, password_hash, kdf_salt, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		user.ID,
		user.Login,
		user.PasswordHash,
		user.KDFSalt,
		user.CreatedAt,
	)
	return err
}

func (r *userRepository) GetByLogin(ctx context.Context, login string) (models.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, login, password_hash, kdf_salt, created_at FROM users WHERE login = $1`,
		login,
	)
	var user models.User
	if err := row.Scan(&user.ID, &user.Login, &user.PasswordHash, &user.KDFSalt, &user.CreatedAt); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, login, password_hash, kdf_salt, created_at FROM users WHERE id = $1`,
		id,
	)
	var user models.User
	if err := row.Scan(&user.ID, &user.Login, &user.PasswordHash, &user.KDFSalt, &user.CreatedAt); err != nil {
		return models.User{}, err
	}
	return user, nil
}
