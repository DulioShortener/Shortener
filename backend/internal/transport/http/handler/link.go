package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/DulioShortener/Shortener/backend/internal/apierrors"
	"github.com/DulioShortener/Shortener/backend/internal/entity"
	"github.com/DulioShortener/Shortener/backend/internal/service"
	authmiddleware "github.com/DulioShortener/Shortener/backend/internal/transport/http/middleware"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/request"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/response"
	"github.com/labstack/echo/v5"
)

type LinkService interface {
	Create(ctx context.Context, userID int64, targetURL string) (entity.Link, error)
	List(ctx context.Context, userID int64) ([]entity.Link, error)
	Delete(ctx context.Context, id, userID int64) error
	Resolve(ctx context.Context, code string) (entity.Link, error)
}

type Link struct {
	service      LinkService
	shortURLBase string
}

func NewLink(linkService LinkService, shortURLBase string) *Link {
	return &Link{service: linkService, shortURLBase: shortURLBase}
}

func (h *Link) Create(c *echo.Context) error {
	principal, ok := authmiddleware.Principal(c.Request().Context())
	if !ok {
		return service.ErrUnauthorized
	}
	var input request.LinkCreate
	if err := bindAndValidate(c, &input); err != nil {
		return err
	}
	created, err := h.service.Create(c.Request().Context(), principal.UserID, input.URL)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, response.LinkFromEntity(created, h.shortURLBase))
}

func (h *Link) List(c *echo.Context) error {
	principal, ok := authmiddleware.Principal(c.Request().Context())
	if !ok {
		return service.ErrUnauthorized
	}
	links, err := h.service.List(c.Request().Context(), principal.UserID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.LinksFromEntities(links, h.shortURLBase))
}

func (h *Link) Delete(c *echo.Context) error {
	principal, ok := authmiddleware.Principal(c.Request().Context())
	if !ok {
		return service.ErrUnauthorized
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return apierrors.BadRequest("INVALID_LINK_ID", "Link ID must be a positive decimal string")
	}
	if err := h.service.Delete(c.Request().Context(), id, principal.UserID); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Link) Resolve(c *echo.Context) error {
	link, err := h.service.Resolve(c.Request().Context(), c.Param("code"))
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, link.TargetURL)
}
