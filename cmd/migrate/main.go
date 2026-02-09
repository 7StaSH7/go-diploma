package main

import (
	"errors"
	"flag"
	"log"

	"github.com/7StaSH7/practicum-diploma/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		up    bool
		down  bool
		steps int
	)

	flag.BoolVar(&up, "up", false, "Run up migrations")
	flag.BoolVar(&down, "down", false, "Run down migrations")
	flag.IntVar(&steps, "steps", 0, "Number of migration steps")

	cfg := config.Load()

	if cfg.POSTGRES_DSN == "" {
		log.Fatal("DSN is required")
	}

	m, err := migrate.New("file://"+cfg.MigrationsPath, cfg.POSTGRES_DSN)
	if err != nil {
		log.Fatalf("migration init failed: %v", err)
	}
	defer m.Close()

	if !up && !down {
		up = true
	}

	if up {
		err = runUp(m, steps)
	} else {
		err = runDown(m, steps)
	}
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

func runUp(m *migrate.Migrate, steps int) error {
	var err error
	if steps > 0 {
		err = m.Steps(steps)
	} else {
		err = m.Up()
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func runDown(m *migrate.Migrate, steps int) error {
	var err error
	if steps > 0 {
		err = m.Steps(-steps)
	} else {
		err = m.Down()
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
