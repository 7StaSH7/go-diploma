package db

import (
	"context"
	"database/sql"

	"github.com/7StaSH7/practicum-diploma/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewDB(cfg config.Config, log *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.POSTGRES_DSN)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func RegisterLifecycle(lc fx.Lifecycle, db *sql.DB, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := db.PingContext(ctx); err != nil {
				log.Error("database ping failed", zap.Error(err))
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return db.Close()
		},
	})
}
