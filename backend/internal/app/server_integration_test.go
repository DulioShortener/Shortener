package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/DulioShortener/Shortener/backend/internal/app"
	"github.com/DulioShortener/Shortener/backend/internal/config"
	"github.com/DulioShortener/Shortener/database"
	"github.com/DulioShortener/Shortener/database/migration"
)

func TestAccountAuthenticationAndLinkLifecycle(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "e2e.db")
	db, err := database.Open(ctx, databasePath)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := migration.NewProvider(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	server, err := app.New(ctx, config.Config{
		DatabasePath: databasePath, BaseHostname: "3dreamstudio.com.br",
		TokenTTL: time.Hour, ShutdownTimeout: time.Second, SonyflakeMachineID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	root := performJSON(t, server.Handler(), http.MethodGet, "/", "", nil)
	if root.Code != http.StatusMovedPermanently || root.Header().Get("Location") != "https://app.3dreamstudio.com.br/" {
		t.Fatalf("root redirect status=%d location=%q", root.Code, root.Header().Get("Location"))
	}
	head := performJSON(t, server.Handler(), http.MethodHead, "/", "", nil)
	if head.Code != http.StatusMovedPermanently || head.Header().Get("Location") != "https://app.3dreamstudio.com.br/" {
		t.Fatalf("root HEAD redirect status=%d location=%q", head.Code, head.Header().Get("Location"))
	}

	preflightRequest := httptest.NewRequest(http.MethodOptions, "/api/v1/links", nil)
	preflightRequest.Header.Set("Origin", "https://unrelated.example")
	preflightRequest.Header.Set("Access-Control-Request-Method", http.MethodGet)
	preflightRequest.Header.Set("Access-Control-Request-Headers", "Authorization,Content-Type")
	preflight := httptest.NewRecorder()
	server.Handler().ServeHTTP(preflight, preflightRequest)
	if preflight.Code != http.StatusNoContent {
		t.Fatalf("CORS preflight status=%d body=%s", preflight.Code, preflight.Body.String())
	}
	if preflight.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("CORS allow origin=%q", preflight.Header().Get("Access-Control-Allow-Origin"))
	}
	if allowedHeaders := preflight.Header().Get("Access-Control-Allow-Headers"); allowedHeaders != "Authorization,Content-Type" {
		t.Fatalf("CORS allow headers=%q", allowedHeaders)
	}
	if credentials := preflight.Header().Get("Access-Control-Allow-Credentials"); credentials != "" {
		t.Fatalf("CORS unexpectedly allows credentials: %q", credentials)
	}

	unknown := performJSON(t, server.Handler(), http.MethodPost, "/api/v1/not-real", "", nil)
	assertAPIError(t, unknown, http.StatusNotFound, "ROUTE_NOT_FOUND")
	wrongMethod := performJSON(t, server.Handler(), http.MethodGet, "/api/v1/auth/login", "", nil)
	assertAPIError(t, wrongMethod, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
	legacyLogin := performJSON(t, server.Handler(), http.MethodPost, "/api/v1/login", "", nil)
	assertAPIError(t, legacyLogin, http.StatusNotFound, "ROUTE_NOT_FOUND")

	invalid := performJSON(t, server.Handler(), http.MethodPost, "/api/v1/users", "", map[string]any{
		"username": "dulio", "password": "password",
	})
	if invalid.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid form status=%d body=%s", invalid.Code, invalid.Body.String())
	}
	var invalidBody struct {
		Errors struct {
			Code   string            `json:"code"`
			Fields map[string]string `json:"fields"`
		} `json:"errors"`
	}
	decode(t, invalid.Body, &invalidBody)
	if invalidBody.Errors.Code != "INVALID_FORM_BODY" || invalidBody.Errors.Fields["password"] == "" {
		t.Fatalf("unexpected validation response: %#v", invalidBody)
	}

	uppercaseUsername := performJSON(t, server.Handler(), http.MethodPost, "/api/v1/users", "", map[string]any{
		"username": "Dulio.User", "password": "Password1!",
	})
	assertAPIError(t, uppercaseUsername, http.StatusUnprocessableEntity, "INVALID_FORM_BODY")

	register := performJSON(t, server.Handler(), http.MethodPost, "/api/v1/users", "", map[string]any{
		"username": "dulio.user", "password": "Password1!",
	})
	if register.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", register.Code, register.Body.String())
	}
	var registered map[string]any
	decode(t, register.Body, &registered)
	if _, ok := registered["id"].(string); !ok {
		t.Fatalf("expected string user id, got %#v", registered["id"])
	}
	if registered["username"] != "dulio.user" || registered["display_name"] != nil {
		t.Fatalf("unexpected registered user: %#v", registered)
	}
	if _, err := time.Parse(time.RFC3339Nano, registered["created_at"].(string)); err != nil {
		t.Fatalf("created_at is not RFC3339: %v", err)
	}

	login := performJSON(t, server.Handler(), http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"username": "dulio.user", "password": "Password1!",
	})
	if login.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", login.Code, login.Body.String())
	}
	var loginBody map[string]any
	decode(t, login.Body, &loginBody)
	token := loginBody["token"].(string)

	badLogin := performJSON(t, server.Handler(), http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"username": "dulio.user", "password": "wrong",
	})
	if badLogin.Code != http.StatusUnauthorized {
		t.Fatalf("bad login status=%d body=%s", badLogin.Code, badLogin.Body.String())
	}
	withoutToken := performJSON(t, server.Handler(), http.MethodGet, "/api/v1/links", "", nil)
	if withoutToken.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status=%d body=%s", withoutToken.Code, withoutToken.Body.String())
	}

	created := performJSON(t, server.Handler(), http.MethodPost, "/api/v1/links", token, map[string]any{
		"url": "https://Example.com:443",
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create link status=%d body=%s", created.Code, created.Body.String())
	}
	var createdBody map[string]any
	decode(t, created.Body, &createdBody)
	linkID := createdBody["id"].(string)
	code := createdBody["code"].(string)
	if shortURL := createdBody["short_url"]; shortURL != "https://3dreamstudio.com.br/r/"+code {
		t.Fatalf("unexpected short URL: %#v", shortURL)
	}

	duplicate := performJSON(t, server.Handler(), http.MethodPost, "/api/v1/links", token, map[string]any{
		"url": "https://example.com/",
	})
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("duplicate status=%d body=%s", duplicate.Code, duplicate.Body.String())
	}

	listed := performJSON(t, server.Handler(), http.MethodGet, "/api/v1/links", token, nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	var listBody struct {
		Links []json.RawMessage `json:"links"`
	}
	decode(t, listed.Body, &listBody)
	if len(listBody.Links) != 1 {
		t.Fatalf("expected one link, got %d", len(listBody.Links))
	}

	redirect := performJSON(t, server.Handler(), http.MethodGet, "/r/"+code, "", nil)
	if redirect.Code != http.StatusFound || redirect.Header().Get("Location") != "https://Example.com:443" {
		t.Fatalf("unexpected redirect status=%d location=%q", redirect.Code, redirect.Header().Get("Location"))
	}

	deleted := performJSON(t, server.Handler(), http.MethodDelete, "/api/v1/links/"+linkID, token, nil)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", deleted.Code, deleted.Body.String())
	}
	logout := performJSON(t, server.Handler(), http.MethodPost, "/api/v1/auth/logout", token, nil)
	if logout.Code != http.StatusNoContent {
		t.Fatalf("logout status=%d body=%s", logout.Code, logout.Body.String())
	}
	unauthorized := performJSON(t, server.Handler(), http.MethodGet, "/api/v1/users/me", token, nil)
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked token to fail, status=%d body=%s", unauthorized.Code, unauthorized.Body.String())
	}
}

func performJSON(t *testing.T, target http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(encoded)
	}
	request := httptest.NewRequest(method, path, reader)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	target.ServeHTTP(recorder, request)
	return recorder
}

func decode(t *testing.T, source io.Reader, destination any) {
	t.Helper()
	if err := json.NewDecoder(source).Decode(destination); err != nil {
		t.Fatal(err)
	}
}

func assertAPIError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Errors struct {
			Code string `json:"code"`
		} `json:"errors"`
	}
	decode(t, response.Body, &body)
	if body.Errors.Code != code {
		t.Fatalf("error code=%q body=%#v", body.Errors.Code, body)
	}
}
