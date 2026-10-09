package handler

import (
	"github.com/DulioShortener/Shortener/backend/internal/apierrors"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/validation"
	"github.com/labstack/echo/v5"
)

func bindAndValidate[T any](c *echo.Context, destination *T) error {
	if err := c.Bind(destination); err != nil {
		return apierrors.MalformedBody()
	}
	if err := c.Validate(destination); err != nil {
		return validation.InvalidForm(err)
	}
	return nil
}
