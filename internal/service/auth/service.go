package auth

//go:generate go run go.uber.org/mock/mockgen@latest -destination=./mocks/auth_repository_mock.go -package=mocks github.com/7StaSH7/practicum-diploma/internal/repository/auth UserRepository,TokenRepository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/7StaSH7/practicum-diploma/internal/config"
	dtoauth "github.com/7StaSH7/practicum-diploma/internal/dto/auth"
	"github.com/7StaSH7/practicum-diploma/internal/models"
	authrepository "github.com/7StaSH7/practicum-diploma/internal/repository/auth"
	"github.com/7StaSH7/practicum-diploma/internal/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Service interface {
	Signup(ctx context.Context, login, password string) (dtoauth.AuthResult, error)
	Signin(ctx context.Context, login, password string) (dtoauth.AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (dtoauth.AuthResult, error)
}

type service struct {
	users  authrepository.UserRepository
	tokens authrepository.TokenRepository
	cfg    config.Config
	log    *zap.Logger
}

func NewService(
	users authrepository.UserRepository,
	tokens authrepository.TokenRepository,
	cfg config.Config,
	log *zap.Logger,
) Service {
	return &service{
		users:  users,
		tokens: tokens,
		cfg:    cfg,
		log:    log,
	}
}

func (s *service) Signup(ctx context.Context, login, password string) (dtoauth.AuthResult, error) {
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return dtoauth.AuthResult{}, err
	}
	kdfSalt, err := utils.NewSalt(16)
	if err != nil {
		return dtoauth.AuthResult{}, err
	}
	user := models.User{
		ID:           uuid.New(),
		Login:        login,
		PasswordHash: passwordHash,
		KDFSalt:      kdfSalt,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.users.Create(ctx, user); err != nil {
		return dtoauth.AuthResult{}, err
	}
	return s.issueTokens(ctx, user)
}

func (s *service) Signin(ctx context.Context, login, password string) (dtoauth.AuthResult, error) {
	user, err := s.users.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dtoauth.AuthResult{}, ErrInvalidCredentials
		}
		return dtoauth.AuthResult{}, err
	}
	valid, err := utils.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return dtoauth.AuthResult{}, err
	}
	if !valid {
		return dtoauth.AuthResult{}, ErrInvalidCredentials
	}
	return s.issueTokens(ctx, user)
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (dtoauth.AuthResult, error) {
	now := time.Now().UTC()
	nextToken, err := utils.NewToken(32)
	if err != nil {
		return dtoauth.AuthResult{}, err
	}
	next := models.RefreshToken{
		ID:        uuid.New(),
		TokenHash: utils.HashToken(nextToken),
		ExpiresAt: now.Add(s.cfg.RefreshTTL),
		CreatedAt: now,
	}

	hash := utils.HashToken(refreshToken)
	rotated, err := s.tokens.Rotate(ctx, hash, now, next)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dtoauth.AuthResult{}, ErrInvalidCredentials
		}
		return dtoauth.AuthResult{}, err
	}

	accessToken, err := utils.NewAccessToken(rotated.UserID.String(), []byte(s.cfg.JWTSecret), s.cfg.AccessTTL)
	if err != nil {
		return dtoauth.AuthResult{}, err
	}
	user, err := s.users.GetByID(ctx, rotated.UserID)
	if err != nil {
		return dtoauth.AuthResult{}, err
	}
	return dtoauth.AuthResult{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: nextToken,
		KDFSalt:      user.KDFSalt,
	}, nil
}

func (s *service) issueTokens(ctx context.Context, user models.User) (dtoauth.AuthResult, error) {
	accessToken, err := utils.NewAccessToken(user.ID.String(), []byte(s.cfg.JWTSecret), s.cfg.AccessTTL)
	if err != nil {
		return dtoauth.AuthResult{}, err
	}
	refreshToken, err := utils.NewToken(32)
	if err != nil {
		return dtoauth.AuthResult{}, err
	}
	refreshHash := utils.HashToken(refreshToken)
	refresh := models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().UTC().Add(s.cfg.RefreshTTL),
		CreatedAt: time.Now().UTC(),
	}
	if err := s.tokens.Create(ctx, refresh); err != nil {
		return dtoauth.AuthResult{}, err
	}
	return dtoauth.AuthResult{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		KDFSalt:      user.KDFSalt,
	}, nil
}
