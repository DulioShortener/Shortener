package validation_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DulioShortener/Shortener/backend/internal/transport/http/request"
	"github.com/DulioShortener/Shortener/backend/internal/transport/http/validation"
)

func TestReviewRatingValidation(t *testing.T) {
	requestValidator, err := validation.New()
	if err != nil {
		t.Fatal(err)
	}
	for _, rating := range []int{-1, 0, 1, 2, 3, 4, 5, 6} {
		t.Run(fmt.Sprintf("rating=%d", rating), func(t *testing.T) {
			err := requestValidator.Validate(request.ReviewCreate{Body: "A review", Rating: &rating})
			if rating >= 1 && rating <= 5 {
				if err != nil {
					t.Fatalf("valid rating rejected: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected rating range error")
			}
			if message := validation.InvalidForm(err).Fields["rating"]; message != "This value must be in range of 1 and 5" {
				t.Fatalf("unexpected range message: %q", message)
			}
		})
	}
	if err := requestValidator.Validate(request.ReviewCreate{Body: "A review"}); err == nil {
		t.Fatal("missing rating accepted")
	} else if message := validation.InvalidForm(err).Fields["rating"]; message != "This value is required" {
		t.Fatalf("unexpected missing rating message: %q", message)
	}
}

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
	for _, username := range []string{"a", "Alicia", "two..dots", "invalid-name", "spaces are bad"} {
		err := requestValidator.Validate(request.UserCreate{Username: username, Password: "Password1!"})
		if err == nil {
			t.Errorf("expected username %q to be rejected", username)
		}
	}
	if err := requestValidator.Validate(request.Login{Username: "Alicia", Password: "Password1!"}); err == nil {
		t.Error("expected uppercase login username to be rejected")
	}
	for _, displayName := range []string{" Alicia", "Alicia ", " a ", "  "} {
		err := requestValidator.Validate(request.UserCreate{
			Username: "alicia", DisplayName: &displayName, Password: "Password1!",
		})
		if err == nil {
			t.Errorf("expected display name %q to be rejected", displayName)
			continue
		}
		apiError := validation.InvalidForm(err)
		if apiError.Fields["display_name"] == "" {
			t.Errorf("expected a display_name field error for %q", displayName)
		}
	}
	for _, targetURL := range []string{"ftp://example.com", "javascript:alert(1)", "https://user:pass@example.com", "https://" + strings.Repeat("a", 4096)} {
		if err := requestValidator.Validate(request.LinkCreate{URL: targetURL}); err == nil {
			t.Errorf("expected URL to be rejected: %.50q", targetURL)
		}
	}
}
