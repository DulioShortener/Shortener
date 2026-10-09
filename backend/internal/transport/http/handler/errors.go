package handler

import (
	"log/slog"
	"net/http"

	"github.com/DulioShortener/Shortener/backend/internal/apierrors"
	"github.com/labstack/echo/v5"
)

func Error(c *echo.Context, err error) {
	if response, _ := echo.UnwrapResponse(c.Response()); response != nil && response.Committed {
		return
	}

	apiError := apierrors.From(err)
	if status := echo.StatusCode(err); status != 0 && apiError.Status == http.StatusInternalServerError {
		apiError = &apierrors.Error{
			Status: status, Code: httpCode(status),
			Message: http.StatusText(status), Cause: err,
		}
	}
	if apiError.Status >= http.StatusInternalServerError {
		slog.Error("request failed", "method", c.Request().Method, "path", c.Request().URL.Path, "error", err)
	}
	if writeErr := c.JSON(apiError.Status, apierrors.ResponseFor(apiError)); writeErr != nil {
		slog.Error("write error response", "error", writeErr)
	}
}

func httpCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "ROUTE_NOT_FOUND"
	case http.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case http.StatusRequestTimeout:
		return "REQUEST_TIMEOUT"
	case http.StatusRequestEntityTooLarge:
		return "REQUEST_BODY_TOO_LARGE"
	case http.StatusUnsupportedMediaType:
		return "UNSUPPORTED_MEDIA_TYPE"
	case http.StatusTooManyRequests:
		return "RATE_LIMIT_EXCEEDED"
	case http.StatusBadGateway:
		return "BAD_GATEWAY"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	default:
		return "HTTP_ERROR"
	}
}
