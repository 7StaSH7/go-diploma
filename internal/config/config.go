package config

import (
	"errors"
	"flag"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ServerURL      string
	POSTGRES_DSN   string
	JWTSecret      string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	MigrationsPath string
}

func Load() (Config, error) {
	v := viper.New()
	v.SetDefault("SERVER_URL", "http://localhost:8080")
	v.SetDefault("POSTGRES_DSN", "")
	v.SetDefault("JWT_SECRET", "change-me")
	v.SetDefault("ACCESS_TTL", 15*time.Minute)
	v.SetDefault("REFRESH_TTL", 7*24*time.Hour)
	v.SetDefault("MIGRATIONS_PATH", "migrations")
	v.SetConfigFile(".env")
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) && !os.IsNotExist(err) {
			return Config{}, err
		}
	}
	v.AutomaticEnv()

	cfg := Config{
		ServerURL:      v.GetString("SERVER_URL"),
		POSTGRES_DSN:   v.GetString("POSTGRES_DSN"),
		JWTSecret:      v.GetString("JWT_SECRET"),
		AccessTTL:      v.GetDuration("ACCESS_TTL"),
		RefreshTTL:     v.GetDuration("REFRESH_TTL"),
		MigrationsPath: v.GetString("MIGRATIONS_PATH"),
	}
	return cfg, nil
}

func BindFlags(fs *flag.FlagSet, cfg *Config) {
	if fs == nil || cfg == nil {
		return
	}
	fs.StringVar(&cfg.ServerURL, "server-url", cfg.ServerURL, "Server base URL")
	fs.StringVar(&cfg.POSTGRES_DSN, "dsn", cfg.POSTGRES_DSN, "PostgreSQL DSN")
	fs.StringVar(&cfg.JWTSecret, "jwt-secret", cfg.JWTSecret, "JWT signing secret")
	fs.DurationVar(&cfg.AccessTTL, "access-ttl", cfg.AccessTTL, "JWT access token TTL")
	fs.DurationVar(&cfg.RefreshTTL, "refresh-ttl", cfg.RefreshTTL, "Refresh token TTL")
	fs.StringVar(&cfg.MigrationsPath, "migrations-path", cfg.MigrationsPath, "Migrations directory")
}

func ResolveHTTPAddr(serverURL string) string {
	parsedURL, err := url.Parse(strings.TrimSpace(serverURL))
	if err != nil || parsedURL.Host == "" {
		return ":8080"
	}
	host := parsedURL.Host
	if !strings.Contains(host, ":") {
		if parsedURL.Scheme == "https" {
			return host + ":443"
		}
		return host + ":80"
	}
	return host
}
