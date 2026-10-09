package apierrors

import (
	"errors"
	"net/http"

	"github.com/DulioShortener/Shortener/backend/internal/service"
)

type Error struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	Cause   error
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

func (e *Error) Unwrap() error {
	return e.Cause
}

type Response struct {
	Errors ErrorBody `json:"errors"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func InvalidForm(fields map[string]string) *Error {
	return &Error{Status: http.StatusUnprocessableEntity, Code: "INVALID_FORM_BODY", Fields: fields}
}

func MalformedBody() *Error {
	return InvalidForm(map[string]string{"body": "The request body must contain valid JSON"})
}

func BadRequest(code, message string) *Error {
	return &Error{Status: http.StatusBadRequest, Code: code, Message: message}
}

func From(err error) *Error {
	var apiError *Error
	if errors.As(err, &apiError) {
		return apiError
	}

	switch {
	case errors.Is(err, service.ErrUsernameTaken):
		return &Error{
			Status: http.StatusConflict, Code: "USERNAME_TAKEN",
			Fields: map[string]string{"username": "This username is already taken"}, Cause: err,
		}
	case errors.Is(err, service.ErrInvalidCredentials):
		return &Error{Status: http.StatusUnauthorized, Code: "INVALID_CREDENTIALS", Message: "Invalid username or password", Cause: err}
	case errors.Is(err, service.ErrUnauthorized):
		return &Error{Status: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: "Authentication is required", Cause: err}
	case errors.Is(err, service.ErrUserNotFound):
		return &Error{Status: http.StatusNotFound, Code: "USER_NOT_FOUND", Message: "User not found", Cause: err}
	case errors.Is(err, service.ErrLinkAlreadyExists):
		return &Error{
			Status: http.StatusConflict, Code: "LINK_ALREADY_EXISTS",
			Fields: map[string]string{"url": "You have already shortened this URL"}, Cause: err,
		}
	case errors.Is(err, service.ErrLinkLimitReached):
		return &Error{Status: http.StatusConflict, Code: "LINK_LIMIT_REACHED", Message: "An account can have at most 50 links", Cause: err}
	case errors.Is(err, service.ErrLinkNotFound):
		return &Error{Status: http.StatusNotFound, Code: "LINK_NOT_FOUND", Message: "Link not found", Cause: err}
	default:
		return &Error{Status: http.StatusInternalServerError, Code: "INTERNAL_SERVER_ERROR", Message: "An unexpected error occurred", Cause: err}
	}
}

func ResponseFor(err *Error) Response {
	return Response{Errors: ErrorBody{Code: err.Code, Message: err.Message, Fields: err.Fields}}
}
