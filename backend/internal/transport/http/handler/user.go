package handler

import (
	"context"
	"net/http"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
	"github.com/DulioShortener/Shortener/backend/internal/service"
	authmiddleware "github.com/DulioShortener/Shortener/backend/internal/transport/http/middleware"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/request"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/response"
	"github.com/labstack/echo/v5"
)

type UserService interface {
	Create(ctx context.Context, input service.UserCreateInput) (entity.User, error)
	Get(ctx context.Context, id int64) (entity.User, error)
}

type User struct {
	service UserService
}

func NewUser(userService UserService) *User {
	return &User{service: userService}
}

func (h *User) Create(c *echo.Context) error {
	var input request.UserCreate
	if err := bindAndValidate(c, &input); err != nil {
		return err
	}
	created, err := h.service.Create(c.Request().Context(), service.UserCreateInput{
		Username: input.Username, DisplayName: input.DisplayName, Password: input.Password,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, response.UserFromEntity(created))
}

func (h *User) Me(c *echo.Context) error {
	principal, ok := authmiddleware.Principal(c.Request().Context())
	if !ok {
		return service.ErrUnauthorized
	}
	user, err := h.service.Get(c.Request().Context(), principal.UserID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.UserFromEntity(user))
}
