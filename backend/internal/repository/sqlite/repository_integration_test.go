package sqlite_test

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
	repository "github.com/DulioShortener/Shortener/backend/internal/repository/sqlite"
	"github.com/DulioShortener/Shortener/backend/internal/service"
	"github.com/DulioShortener/Shortener/database"
	"github.com/DulioShortener/Shortener/database/migration"
)

func TestRepositoriesEnforceOwnershipDuplicatesAndLimit(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "repository.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	provider, err := migration.NewProvider(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}

	users := repository.NewUserRepository(db)
	links := repository.NewLinkRepository(db)
	now := time.Now().UTC().Truncate(time.Millisecond)
	user, err := users.Create(ctx, entity.User{
		ID: 1, Username: "owner", PasswordHash: "hash", CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := users.Create(ctx, entity.User{
		ID: 2, Username: "owner", PasswordHash: "hash", CreatedAt: now, UpdatedAt: now,
	}); !errors.Is(err, service.ErrUsernameTaken) {
		t.Fatalf("expected case-insensitive duplicate username, got %v", err)
	}

	first := entity.Link{
		ID: 100, UserID: user.ID, Code: "00000000", TargetURL: "https://example.com",
		TargetURLKey: "https://example.com/", CreatedAt: now,
	}
	if _, err := links.Create(ctx, first); err != nil {
		t.Fatal(err)
	}
	duplicate := first
	duplicate.ID = 101
	duplicate.Code = "00000001"
	if _, err := links.Create(ctx, duplicate); !errors.Is(err, service.ErrLinkAlreadyExists) {
		t.Fatalf("expected duplicate URL error, got %v", err)
	}

	for index := 1; index < 50; index++ {
		link := entity.Link{
			ID: int64(101 + index), UserID: user.ID, Code: fmt.Sprintf("%08d", index+1),
			TargetURL:    fmt.Sprintf("https://example.com/%d", index),
			TargetURLKey: fmt.Sprintf("https://example.com/%d", index), CreatedAt: now.Add(time.Duration(index)),
		}
		if _, err := links.Create(ctx, link); err != nil {
			t.Fatalf("create link %d: %v", index, err)
		}
	}
	overLimit := entity.Link{
		ID: 999, UserID: user.ID, Code: "99999999", TargetURL: "https://limit.example",
		TargetURLKey: "https://limit.example/", CreatedAt: now,
	}
	if _, err := links.Create(ctx, overLimit); !errors.Is(err, service.ErrLinkLimitReached) {
		t.Fatalf("expected link limit error, got %v", err)
	}

	listed, err := links.ListByUser(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 50 {
		t.Fatalf("expected 50 links, got %d", len(listed))
	}
	if err := links.Delete(ctx, first.ID, 999); !errors.Is(err, service.ErrLinkNotFound) {
		t.Fatalf("expected ownership-safe not found, got %v", err)
	}
}
