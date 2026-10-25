package sqlite_test

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
	repository "github.com/DulioShortener/Shortener/backend/internal/repository/sqlite"
	"github.com/DulioShortener/Shortener/backend/internal/service"
	"github.com/DulioShortener/Shortener/database"
	"github.com/DulioShortener/Shortener/database/migration"
)

func TestReviewRepositoryLifecycleAndPagination(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "reviews.db"))
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
	reviews := repository.NewReviewRepository(db)
	now := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
	for userID := int64(1); userID <= 53; userID++ {
		_, err := users.Create(ctx, entity.User{
			ID: userID, Username: fmt.Sprintf("reviewer%d", userID),
			PasswordHash: "hash", CreatedAt: now, UpdatedAt: now,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	createRoot := func(id, userID int64, at time.Time) entity.ReviewComment {
		t.Helper()
		rating := 5
		created, err := reviews.CreateRoot(ctx, entity.ReviewComment{
			ID: id, UserID: userID, Body: "Excellent", Rating: &rating, CreatedAt: at,
		})
		if err != nil {
			t.Fatal(err)
		}
		return created
	}
	root := createRoot(100, 1, now)
	if _, err := reviews.CreateRoot(ctx, entity.ReviewComment{
		ID: 101, UserID: 1, Body: "Again", Rating: root.Rating, CreatedAt: now,
	}); !errors.Is(err, service.ErrReviewAlreadyExists) {
		t.Fatalf("expected active-review conflict, got %v", err)
	}
	for index := int64(2); index <= 52; index++ {
		createRoot(100+index, index, now.Add(time.Duration(index)*time.Millisecond))
	}
	firstPage, err := reviews.ListRoots(ctx, nil, false)
	if err != nil || len(firstPage.Page.Items) != 50 || !firstPage.Page.HasMore || firstPage.Page.NextCursor == nil {
		t.Fatalf("first root page: count=%d more=%t err=%v", len(firstPage.Page.Items), firstPage.Page.HasMore, err)
	}
	secondPage, err := reviews.ListRoots(ctx, firstPage.Page.NextCursor, false)
	if err != nil || len(secondPage.Page.Items) != 2 || secondPage.Page.HasMore {
		t.Fatalf("second root page: count=%d more=%t err=%v", len(secondPage.Page.Items), secondPage.Page.HasMore, err)
	}
	if secondPage.Page.Items[1].ID != root.ID {
		t.Fatalf("root cursor skipped oldest review: %+v", secondPage.Page.Items)
	}

	reply := func(id, parentID int64, at time.Time) entity.ReviewComment {
		t.Helper()
		created, err := reviews.CreateReply(ctx, entity.ReviewComment{
			ID: id, UserID: 2, ParentID: &parentID, Body: "Reply", CreatedAt: at,
		})
		if err != nil {
			t.Fatal(err)
		}
		return created
	}
	firstReply := reply(1000, root.ID, now)
	deepReply := reply(1001, firstReply.ID, now)
	for index := int64(0); index < 50; index++ {
		reply(1100+index, root.ID, now.Add(time.Duration(index+1)*time.Millisecond))
	}
	preview, err := reviews.Preview(ctx, root.ID)
	if err != nil || len(preview.Items) != 1 || preview.Items[0].ID != firstReply.ID || !preview.HasMore {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
	withPreviews, err := reviews.ListRoots(ctx, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, present := withPreviews.Previews[root.ID]; present {
		t.Fatal("the oldest root appeared in the first page previews")
	}
	oldestWithPreviews, err := reviews.ListRoots(ctx, firstPage.Page.NextCursor, true)
	if err != nil || len(oldestWithPreviews.Previews[root.ID].Items) != 1 ||
		!oldestWithPreviews.Previews[root.ID].HasMore {
		t.Fatalf("batched root preview=%+v err=%v", oldestWithPreviews.Previews[root.ID], err)
	}
	children, err := reviews.ListReplies(ctx, root.ID, nil)
	if err != nil || len(children.Items) != 50 || !children.HasMore {
		t.Fatalf("children=%d more=%t err=%v", len(children.Items), children.HasMore, err)
	}
	lastChildPage, err := reviews.ListReplies(ctx, root.ID, children.NextCursor)
	if err != nil || len(lastChildPage.Items) != 1 || lastChildPage.HasMore {
		t.Fatalf("last children=%+v err=%v", lastChildPage, err)
	}
	if !children.Items[0].HasReplies || children.Items[0].ID != firstReply.ID {
		t.Fatalf("first reply lost deep-child flag: %+v", children.Items[0])
	}
	if err := reviews.SoftDelete(ctx, firstReply.ID, 1, now); !errors.Is(err, service.ErrReviewNotFound) {
		t.Fatalf("wrong owner delete: %v", err)
	}
	if err := reviews.SoftDelete(ctx, firstReply.ID, 2, now); err != nil {
		t.Fatal(err)
	}
	tombstone, err := reviews.FindVisible(ctx, firstReply.ID)
	if err != nil || tombstone.DeletedAt == nil {
		t.Fatalf("deleted reply should remain as a tombstone: %+v %v", tombstone, err)
	}
	if _, err := reviews.FindVisible(ctx, deepReply.ID); err != nil {
		t.Fatalf("descendant should remain visible: %v", err)
	}
	if _, err := reviews.CreateReply(ctx, entity.ReviewComment{
		ID: 2000, UserID: 2, ParentID: &firstReply.ID, Body: "New", CreatedAt: now,
	}); !errors.Is(err, service.ErrReviewNotFound) {
		t.Fatalf("reply to deleted parent: %v", err)
	}
	if err := reviews.SoftDelete(ctx, root.ID, 1, now); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{root.ID, firstReply.ID, deepReply.ID} {
		if _, err := reviews.FindVisible(ctx, id); !errors.Is(err, service.ErrReviewNotFound) {
			t.Fatalf("deleted root left %d visible: %v", id, err)
		}
	}
	if _, err := reviews.ListReplies(ctx, root.ID, nil); !errors.Is(err, service.ErrReviewNotFound) {
		t.Fatalf("deleted root replies: %v", err)
	}
	createRoot(3000, 1, now.Add(time.Hour))
}

func TestConcurrentReviewsAllowOneActiveRoot(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "concurrent-reviews.db"))
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
	now := time.Now().UTC().Truncate(time.Millisecond)
	users := repository.NewUserRepository(db)
	if _, err := users.Create(ctx, entity.User{
		ID: 1, Username: "concurrent", PasswordHash: "hash", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	reviews := repository.NewReviewRepository(db)
	start := make(chan struct{})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for _, id := range []int64{10, 11} {
		workers.Add(1)
		go func(id int64) {
			defer workers.Done()
			<-start
			rating := 5
			_, err := reviews.CreateRoot(ctx, entity.ReviewComment{
				ID: id, UserID: 1, Body: "Once", Rating: &rating, CreatedAt: now,
			})
			results <- err
		}(id)
	}
	close(start)
	workers.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, service.ErrReviewAlreadyExists):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent insert error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
}
