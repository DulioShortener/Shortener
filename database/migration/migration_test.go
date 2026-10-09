package migration_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/DulioShortener/Shortener/database"
	"github.com/DulioShortener/Shortener/database/migration"
)

func TestMigrationsApplyAndRollback(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "migration.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	provider, err := migration.NewProvider(db)
	if err != nil {
		t.Fatal(err)
	}
	results, err := provider.Up(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 migrations, got %d", len(results))
	}

	for _, name := range []string{"users", "auth_sessions", "links"} {
		var count int
		if err := db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?
		`, name).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("expected table %s to exist", name)
		}
	}

	for range 3 {
		if _, err := provider.Down(ctx); err != nil {
			t.Fatal(err)
		}
	}
	version, err := provider.GetDBVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != 0 {
		t.Fatalf("expected version 0, got %d", version)
	}
}
