package main

import (
	"context"
	"path/filepath"
	"testing"
)

func TestRunSupportsIndependentMigrationCommands(t *testing.T) {
	t.Setenv("DATABASE_PATH", filepath.Join(t.TempDir(), "cli.db"))
	ctx := context.Background()

	for _, command := range []string{"up", "status", "version", "down"} {
		if err := run(ctx, []string{command}); err != nil {
			t.Fatalf("command %s failed: %v", command, err)
		}
	}
	if err := run(ctx, []string{"unknown"}); err == nil {
		t.Fatal("expected unknown command to fail")
	}
}
