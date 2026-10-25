package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
)

type ReviewListing struct {
	Page     entity.ReviewPage
	Previews map[int64]entity.ReviewPage
}

type ReviewDetail struct {
	Comment entity.ReviewComment
	Preview *entity.ReviewPage
}

type ReviewRepository interface {
	CreateRoot(ctx context.Context, comment entity.ReviewComment) (entity.ReviewComment, error)
	CreateReply(ctx context.Context, comment entity.ReviewComment) (entity.ReviewComment, error)
	FindVisible(ctx context.Context, id int64) (entity.ReviewComment, error)
	FindActiveRootByUser(ctx context.Context, userID int64) (entity.ReviewComment, error)
	ListRoots(ctx context.Context, cursor *entity.ReviewCursor, includePreviews bool) (ReviewListing, error)
	ListReplies(ctx context.Context, parentID int64, cursor *entity.ReviewCursor) (entity.ReviewPage, error)
	Preview(ctx context.Context, parentID int64) (entity.ReviewPage, error)
	SoftDelete(ctx context.Context, id, userID int64, deletedAt time.Time) error
}

type ReviewService struct {
	reviews ReviewRepository
	ids     IDGenerator
	clock   Clock
}

func NewReviewService(reviews ReviewRepository, ids IDGenerator, clock Clock) *ReviewService {
	return &ReviewService{reviews: reviews, ids: ids, clock: clock}
}

func (s *ReviewService) CreateRoot(ctx context.Context, userID int64, body string, rating int) (entity.ReviewComment, error) {
	if !validReviewBody(body) || rating < 1 || rating > 5 {
		return entity.ReviewComment{}, ErrInvalidReview
	}
	id, err := s.ids.NextID()
	if err != nil {
		return entity.ReviewComment{}, fmt.Errorf("allocate review ID: %w", err)
	}
	return s.reviews.CreateRoot(ctx, entity.ReviewComment{
		ID: id, UserID: userID, Body: body, Rating: &rating,
		CreatedAt: s.clock.Now().UTC().Truncate(time.Millisecond),
	})
}

func (s *ReviewService) CreateReply(ctx context.Context, userID, parentID int64, body string) (entity.ReviewComment, error) {
	if !validReviewBody(body) {
		return entity.ReviewComment{}, ErrInvalidReview
	}
	id, err := s.ids.NextID()
	if err != nil {
		return entity.ReviewComment{}, fmt.Errorf("allocate reply ID: %w", err)
	}
	return s.reviews.CreateReply(ctx, entity.ReviewComment{
		ID: id, UserID: userID, ParentID: &parentID, Body: body,
		CreatedAt: s.clock.Now().UTC().Truncate(time.Millisecond),
	})
}

func (s *ReviewService) List(ctx context.Context, cursor *entity.ReviewCursor, authenticated bool) (ReviewListing, error) {
	return s.reviews.ListRoots(ctx, cursor, authenticated)
}

func (s *ReviewService) ListReplies(ctx context.Context, parentID int64, cursor *entity.ReviewCursor) (entity.ReviewPage, error) {
	return s.reviews.ListReplies(ctx, parentID, cursor)
}

func (s *ReviewService) Get(ctx context.Context, id int64, authenticated bool) (ReviewDetail, error) {
	comment, err := s.reviews.FindVisible(ctx, id)
	if err != nil {
		return ReviewDetail{}, err
	}
	if comment.ParentID != nil && !authenticated {
		return ReviewDetail{}, ErrUnauthorized
	}
	detail := ReviewDetail{Comment: comment}
	if authenticated && comment.ParentID == nil {
		preview, err := s.reviews.Preview(ctx, id)
		if err != nil {
			return ReviewDetail{}, err
		}
		detail.Preview = &preview
	}
	return detail, nil
}

func (s *ReviewService) Mine(ctx context.Context, userID int64) (ReviewDetail, error) {
	comment, err := s.reviews.FindActiveRootByUser(ctx, userID)
	if err != nil {
		return ReviewDetail{}, err
	}
	preview, err := s.reviews.Preview(ctx, comment.ID)
	if err != nil {
		return ReviewDetail{}, err
	}
	return ReviewDetail{Comment: comment, Preview: &preview}, nil
}

func (s *ReviewService) Delete(ctx context.Context, id, userID int64) error {
	return s.reviews.SoftDelete(ctx, id, userID, s.clock.Now().UTC().Truncate(time.Millisecond))
}

func validReviewBody(body string) bool {
	return strings.TrimSpace(body) != "" && utf8.ValidString(body) && utf8.RuneCountInString(body) <= 2000
}
