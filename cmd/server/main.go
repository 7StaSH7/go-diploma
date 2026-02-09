package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/7StaSH7/practicum-diploma/internal/config"
	"github.com/7StaSH7/practicum-diploma/internal/db"
	authhandler "github.com/7StaSH7/practicum-diploma/internal/handler/auth"
	secrethandler "github.com/7StaSH7/practicum-diploma/internal/handler/secret"
	"github.com/7StaSH7/practicum-diploma/internal/logger"
	authrepository "github.com/7StaSH7/practicum-diploma/internal/repository/auth"
	secretrepository "github.com/7StaSH7/practicum-diploma/internal/repository/secret"
	"github.com/7StaSH7/practicum-diploma/internal/server"
	authservice "github.com/7StaSH7/practicum-diploma/internal/service/auth"
	secretservice "github.com/7StaSH7/practicum-diploma/internal/service/secret"
	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := fx.New(
		fx.Provide(config.Load),
		fx.Provide(logger.New),
		fx.Provide(db.NewDB),
		fx.Invoke(db.RegisterLifecycle),
		fx.Provide(authrepository.NewUserRepository),
		fx.Provide(authrepository.NewTokenRepository),
		fx.Provide(secretrepository.NewSecretRepository),
		fx.Provide(authservice.NewService),
		fx.Provide(secretservice.NewService),
		fx.Provide(authhandler.New),
		fx.Provide(secrethandler.New),
		fx.Provide(server.NewRouter),
		fx.Invoke(server.RegisterRoutes),
		fx.Invoke(server.StartHTTPServer),
	)

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return app.Start(ctx)
	})
	group.Go(func() error {
		<-ctx.Done()
		return app.Stop(context.Background())
	})

	if err := group.Wait(); err != nil {
		log.Printf("server stopped: %v", err)
	}
}
