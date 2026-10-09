package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/DulioShortener/Shortener/backend/internal/app"
	"github.com/DulioShortener/Shortener/backend/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load configuration", "error", err)
		os.Exit(1)
	}
	server, err := app.New(ctx, cfg)
	if err != nil {
		slog.Error("initialize server", "error", err)
		os.Exit(1)
	}
	defer server.Close()

	if err := server.Run(ctx); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
