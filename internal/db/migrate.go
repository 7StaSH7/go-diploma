package db

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/7StaSH7/practicum-diploma/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func RunMigrations(lc fx.Lifecycle, cfg config.Config, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			path := cfg.MigrationsPath
			if !filepath.IsAbs(path) {
				cwd, err := filepath.Abs(".")
				if err == nil {
					path = filepath.Join(cwd, path)
				}
			}
			sourceURL := "file://" + strings.TrimSuffix(path, "/")
			m, err := migrate.New(sourceURL, cfg.POSTGRES_DSN)
			if err != nil {
				log.Error("migration init failed", zap.Error(err))
				return err
			}
			defer m.Close()
			err = m.Up()
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				log.Error("migration failed", zap.Error(err))
				return err
			}
			return nil
		},
	})
}
