package handler

import (
	"context"
	"net/http"

	"github.com/DulioShortener/Shortener/backend/internal/service"
	authmiddleware "github.com/DulioShortener/Shortener/backend/internal/transport/http/middleware"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/request"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/response"
	"github.com/labstack/echo/v5"
)

type AuthService interface {
	Login(ctx context.Context, input service.LoginInput) (service.LoginResult, error)
	Logout(ctx context.Context, principal service.Principal) error
}

type Auth struct {
	service AuthService
}

func NewAuth(authService AuthService) *Auth {
	return &Auth{service: authService}
}

func (h *Auth) Login(c *echo.Context) error {
	var input request.Login
	if err := bindAndValidate(c, &input); err != nil {
		return err
	}
	result, err := h.service.Login(c.Request().Context(), service.LoginInput{
		Username: input.Username, Password: input.Password,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.LoginFromResult(result))
}

func (h *Auth) Logout(c *echo.Context) error {
	principal, ok := authmiddleware.Principal(c.Request().Context())
	if !ok {
		return service.ErrUnauthorized
	}
	if err := h.service.Logout(c.Request().Context(), principal); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
