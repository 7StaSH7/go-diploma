package config

import (
	"flag"
	"testing"
	"time"
)

func TestLoadRespectsEnvOverrides(t *testing.T) {
	t.Setenv("SERVER_URL", "http://env.test")
	t.Setenv("POSTGRES_DSN", "postgres://env")
	t.Setenv("JWT_SECRET", "env-secret")
	t.Setenv("ACCESS_TTL", "1m")
	t.Setenv("REFRESH_TTL", "24h")
	t.Setenv("MIGRATIONS_PATH", "db/migrations")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.ServerURL != "http://env.test" {
		t.Fatalf("unexpected server url: %s", cfg.ServerURL)
	}
	if cfg.POSTGRES_DSN != "postgres://env" {
		t.Fatalf("unexpected dsn: %s", cfg.POSTGRES_DSN)
	}
	if cfg.JWTSecret != "env-secret" {
		t.Fatalf("unexpected jwt secret: %s", cfg.JWTSecret)
	}
	if cfg.AccessTTL != time.Minute {
		t.Fatalf("unexpected access ttl: %v", cfg.AccessTTL)
	}
	if cfg.RefreshTTL != 24*time.Hour {
		t.Fatalf("unexpected refresh ttl: %v", cfg.RefreshTTL)
	}
	if cfg.MigrationsPath != "db/migrations" {
		t.Fatalf("unexpected migrations path: %s", cfg.MigrationsPath)
	}
}

func TestBindFlagsOverridesConfig(t *testing.T) {
	cfg := Config{
		ServerURL:      "http://localhost:8080",
		POSTGRES_DSN:   "postgres://default",
		JWTSecret:      "default-secret",
		AccessTTL:      15 * time.Minute,
		RefreshTTL:     7 * 24 * time.Hour,
		MigrationsPath: "migrations",
	}

	fs := flag.NewFlagSet("config-test", flag.ContinueOnError)
	BindFlags(fs, &cfg)
	err := fs.Parse([]string{
		"--server-url", "http://flag.test",
		"--dsn", "postgres://flag",
		"--jwt-secret", "flag-secret",
		"--access-ttl", "2m",
		"--refresh-ttl", "48h",
		"--migrations-path", "custom/migrations",
	})
	if err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	if cfg.ServerURL != "http://flag.test" {
		t.Fatalf("unexpected server url: %s", cfg.ServerURL)
	}
	if cfg.POSTGRES_DSN != "postgres://flag" {
		t.Fatalf("unexpected dsn: %s", cfg.POSTGRES_DSN)
	}
	if cfg.JWTSecret != "flag-secret" {
		t.Fatalf("unexpected jwt secret: %s", cfg.JWTSecret)
	}
	if cfg.AccessTTL != 2*time.Minute {
		t.Fatalf("unexpected access ttl: %v", cfg.AccessTTL)
	}
	if cfg.RefreshTTL != 48*time.Hour {
		t.Fatalf("unexpected refresh ttl: %v", cfg.RefreshTTL)
	}
	if cfg.MigrationsPath != "custom/migrations" {
		t.Fatalf("unexpected migrations path: %s", cfg.MigrationsPath)
	}
}
