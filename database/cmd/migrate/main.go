package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/DulioShortener/Shortener/database"
	"github.com/DulioShortener/Shortener/database/migration"
)

const defaultDatabasePath = "./data/dulio.db"

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		slog.Error("migration command failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	command := "up"
	if len(args) > 0 {
		command = strings.ToLower(strings.TrimSpace(args[0]))
	}

	databasePath := strings.TrimSpace(os.Getenv("DATABASE_PATH"))
	if databasePath == "" {
		databasePath = defaultDatabasePath
	}

	db, err := database.Open(ctx, databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	provider, err := migration.NewProvider(db)
	if err != nil {
		return err
	}

	switch command {
	case "up":
		results, err := provider.Up(ctx)
		if err != nil {
			return fmt.Errorf("apply migrations: %w", err)
		}
		slog.Info("migrations applied", "count", len(results))
		return nil
	case "down":
		result, err := provider.Down(ctx)
		if err != nil {
			return fmt.Errorf("roll back migration: %w", err)
		}
		slog.Info("migration rolled back", "source", result.Source.Path)
		return nil
	case "status":
		results, err := provider.Status(ctx)
		if err != nil {
			return fmt.Errorf("read migration status: %w", err)
		}
		for _, result := range results {
			fmt.Printf("%s\t%s\n", result.State, result.Source.Path)
		}
		return nil
	case "version":
		version, err := provider.GetDBVersion(ctx)
		if err != nil {
			return fmt.Errorf("read migration version: %w", err)
		}
		fmt.Println(version)
		return nil
	default:
		return fmt.Errorf("unknown command %q; expected up, down, status, or version", command)
	}
}
