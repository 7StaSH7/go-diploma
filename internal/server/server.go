package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/7StaSH7/practicum-diploma/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func StartHTTPServer(lc fx.Lifecycle, cfg config.Config, router *gin.Engine, log *zap.Logger) {
	listenAddr := config.ResolveHTTPAddr(cfg.ServerURL)
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       time.Minute,
	}
	var listener net.Listener

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			var err error
			listener, err = net.Listen("tcp", listenAddr)
			if err != nil {
				return err
			}
			go func() {
				if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Error("http server stopped", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
