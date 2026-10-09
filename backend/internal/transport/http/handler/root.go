package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type Root struct {
	applicationURL string
}

func NewRoot(applicationURL string) *Root {
	return &Root{applicationURL: applicationURL}
}

func (h *Root) Redirect(c *echo.Context) error {
	return c.Redirect(http.StatusMovedPermanently, h.applicationURL)
}
