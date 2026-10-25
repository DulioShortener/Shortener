package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DulioShortener/Shortener/backend/internal/apierrors"
	"github.com/DulioShortener/Shortener/backend/internal/entity"
	"github.com/DulioShortener/Shortener/backend/internal/service"
	authmiddleware "github.com/DulioShortener/Shortener/backend/internal/transport/http/middleware"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/request"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/response"
	"github.com/labstack/echo/v5"
)

type ReviewService interface {
	CreateRoot(ctx context.Context, userID int64, body string, rating int) (entity.ReviewComment, error)
	CreateReply(ctx context.Context, userID, parentID int64, body string) (entity.ReviewComment, error)
	List(ctx context.Context, cursor *entity.ReviewCursor, authenticated bool) (service.ReviewListing, error)
	ListReplies(ctx context.Context, parentID int64, cursor *entity.ReviewCursor) (entity.ReviewPage, error)
	Get(ctx context.Context, id int64, authenticated bool) (service.ReviewDetail, error)
	Mine(ctx context.Context, userID int64) (service.ReviewDetail, error)
	Delete(ctx context.Context, id, userID int64) error
}

type Review struct {
	service ReviewService
}

func NewReview(reviewService ReviewService) *Review {
	return &Review{service: reviewService}
}

func (h *Review) Create(c *echo.Context) error {
	principal, ok := authmiddleware.Principal(c.Request().Context())
	if !ok {
		return service.ErrUnauthorized
	}
	var input request.ReviewCreate
	if err := bindAndValidate(c, &input); err != nil {
		return err
	}
	created, err := h.service.CreateRoot(c.Request().Context(), principal.UserID, input.Body, *input.Rating)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, response.ReviewCommentFromEntity(created, true))
}

func (h *Review) List(c *echo.Context) error {
	cursor, err := reviewCursor(c.QueryParam("cursor"))
	if err != nil {
		return err
	}
	_, authenticated := authmiddleware.Principal(c.Request().Context())
	listing, err := h.service.List(c.Request().Context(), cursor, authenticated)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.ReviewListingFromService(listing, authenticated))
}

func (h *Review) Mine(c *echo.Context) error {
	principal, ok := authmiddleware.Principal(c.Request().Context())
	if !ok {
		return service.ErrUnauthorized
	}
	detail, err := h.service.Mine(c.Request().Context(), principal.UserID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.ReviewDetailFromService(detail, true))
}

func (h *Review) Get(c *echo.Context) error {
	id, err := reviewID(c.Param("id"))
	if err != nil {
		return err
	}
	_, authenticated := authmiddleware.Principal(c.Request().Context())
	detail, err := h.service.Get(c.Request().Context(), id, authenticated)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.ReviewDetailFromService(detail, authenticated))
}

func (h *Review) Delete(c *echo.Context) error {
	principal, ok := authmiddleware.Principal(c.Request().Context())
	if !ok {
		return service.ErrUnauthorized
	}
	id, err := reviewID(c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.service.Delete(c.Request().Context(), id, principal.UserID); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Review) CreateReply(c *echo.Context) error {
	principal, ok := authmiddleware.Principal(c.Request().Context())
	if !ok {
		return service.ErrUnauthorized
	}
	parentID, err := reviewID(c.Param("id"))
	if err != nil {
		return err
	}
	var input request.ReplyCreate
	if err := bindAndValidate(c, &input); err != nil {
		return err
	}
	if len(input.Rating) != 0 {
		return apierrors.InvalidForm(map[string]string{"rating": "Replies cannot have a rating"})
	}
	created, err := h.service.CreateReply(c.Request().Context(), principal.UserID, parentID, input.Body)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, response.ReviewCommentFromEntity(created, true))
}

func (h *Review) ListReplies(c *echo.Context) error {
	parentID, err := reviewID(c.Param("id"))
	if err != nil {
		return err
	}
	cursor, err := reviewCursor(c.QueryParam("cursor"))
	if err != nil {
		return err
	}
	page, err := h.service.ListReplies(c.Request().Context(), parentID, cursor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.ReviewPageFromEntity(page, true))
}

func reviewID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 || strconv.FormatInt(id, 10) != raw {
		return 0, apierrors.BadRequest("INVALID_REVIEW_ID", "Review ID must be a positive decimal string")
	}
	return id, nil
}

func reviewCursor(raw string) (*entity.ReviewCursor, error) {
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ":")
	if len(parts) != 2 {
		return nil, apierrors.BadRequest("INVALID_CURSOR", "Cursor must contain a timestamp and ID")
	}
	milliseconds, timestampError := strconv.ParseInt(parts[0], 10, 64)
	id, idError := reviewID(parts[1])
	if timestampError != nil || milliseconds <= 0 || idError != nil || strconv.FormatInt(milliseconds, 10) != parts[0] {
		return nil, apierrors.BadRequest("INVALID_CURSOR", "Cursor must contain a timestamp and ID")
	}
	return &entity.ReviewCursor{CreatedAt: time.UnixMilli(milliseconds).UTC(), ID: id}, nil
}
