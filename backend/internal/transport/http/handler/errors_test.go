package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DulioShortener/Shortener/backend/internal/apierrors"
	"github.com/labstack/echo/v5"
)

func TestErrorPreservesEchoHTTPStatus(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "not found", err: echo.ErrNotFound, status: http.StatusNotFound, code: "ROUTE_NOT_FOUND"},
		{name: "method not allowed", err: echo.ErrMethodNotAllowed, status: http.StatusMethodNotAllowed, code: "METHOD_NOT_ALLOWED"},
		{name: "body too large", err: echo.ErrStatusRequestEntityTooLarge, status: http.StatusRequestEntityTooLarge, code: "REQUEST_BODY_TOO_LARGE"},
		{name: "unsupported media type", err: echo.ErrUnsupportedMediaType, status: http.StatusUnsupportedMediaType, code: "UNSUPPORTED_MEDIA_TYPE"},
		{name: "rate limited", err: echo.ErrTooManyRequests, status: http.StatusTooManyRequests, code: "RATE_LIMIT_EXCEEDED"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			recorder := httptest.NewRecorder()
			context := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), recorder)

			Error(context, test.err)

			if recorder.Code != test.status {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			var response apierrors.Response
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}
			if response.Errors.Code != test.code || response.Errors.Message != http.StatusText(test.status) {
				t.Fatalf("unexpected response: %#v", response)
			}
		})
	}
}
