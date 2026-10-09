package middleware

import (
	"context"
	"strings"

	"github.com/DulioShortener/Shortener/backend/internal/service"
	"github.com/labstack/echo/v5"
)

type principalContextKey struct{}

type Authenticator interface {
	Authenticate(ctx context.Context, rawToken string) (service.Principal, error)
}

type Authentication struct {
	authenticator Authenticator
}

func NewAuthentication(authenticator Authenticator) *Authentication {
	return &Authentication{authenticator: authenticator}
}

func (m *Authentication) Require(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		token, ok := bearerToken(c.Request().Header.Get("Authorization"))
		if !ok {
			return service.ErrUnauthorized
		}
		principal, err := m.authenticator.Authenticate(c.Request().Context(), token)
		if err != nil {
			return err
		}
		ctx := context.WithValue(c.Request().Context(), principalContextKey{}, principal)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func Principal(ctx context.Context) (service.Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(service.Principal)
	return principal, ok
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
