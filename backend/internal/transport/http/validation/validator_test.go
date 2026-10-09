package validation_test

import (
	"strings"
	"testing"

	"github.com/DulioShortener/Shortener/backend/internal/transport/http/request"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/validation"
)

func TestUserCreateValidation(t *testing.T) {
	requestValidator, err := validation.New()
	if err != nil {
		t.Fatal(err)
	}
	displayName := "Dulio"
	valid := request.UserCreate{
		Username: "dulio.user", DisplayName: &displayName, Password: "Password1!",
	}
	if err := requestValidator.Validate(valid); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	valid.DisplayName = nil
	if err := requestValidator.Validate(valid); err != nil {
		t.Fatalf("optional display name rejected: %v", err)
	}

	tests := []struct {
		name     string
		password string
		tagText  string
	}{
		{name: "too short", password: "Aa1!", tagText: "at least 8"},
		{name: "lowercase", password: "PASSWORD1!", tagText: "lowercase"},
		{name: "uppercase", password: "password1!", tagText: "uppercase"},
		{name: "digit", password: "Password!", tagText: "digit"},
		{name: "special", password: "Password1", tagText: "special"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := request.UserCreate{Username: "valid.user", Password: test.password}
			err := requestValidator.Validate(input)
			if err == nil {
				t.Fatal("expected validation error")
			}
			apiError := validation.InvalidForm(err)
			if !strings.Contains(apiError.Fields["password"], test.tagText) {
				t.Fatalf("unexpected password error: %q", apiError.Fields["password"])
			}
		})
	}
}

func TestUsernameAndURLValidation(t *testing.T) {
	requestValidator, err := validation.New()
	if err != nil {
		t.Fatal(err)
	}
	for _, username := range []string{"a", "two..dots", "invalid-name", "spaces are bad"} {
		err := requestValidator.Validate(request.UserCreate{Username: username, Password: "Password1!"})
		if err == nil {
			t.Errorf("expected username %q to be rejected", username)
		}
	}
	for _, targetURL := range []string{"ftp://example.com", "javascript:alert(1)", "https://user:pass@example.com", "https://" + strings.Repeat("a", 4096)} {
		if err := requestValidator.Validate(request.LinkCreate{URL: targetURL}); err == nil {
			t.Errorf("expected URL to be rejected: %.50q", targetURL)
		}
	}
}
