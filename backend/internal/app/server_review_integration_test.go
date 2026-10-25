package app_test

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DulioShortener/Shortener/backend/internal/app"
	"github.com/DulioShortener/Shortener/backend/internal/config"
	"github.com/DulioShortener/Shortener/database"
	"github.com/DulioShortener/Shortener/database/migration"
)

func TestReviewHTTPVisibilityOwnershipAndSoftDeletion(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "reviews-http.db")
	db, err := database.Open(ctx, path)
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
		DatabasePath: path, BaseHostname: "3dreamstudio.com.br",
		TokenTTL: time.Hour, ShutdownTimeout: time.Second, SonyflakeMachineID: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	target := server.Handler()
	registerAndLogin := func(username string) string {
		t.Helper()
		created := performJSON(t, target, http.MethodPost, "/api/v1/users", "", map[string]any{
			"username": username, "password": "Password1!",
		})
		if created.Code != http.StatusCreated {
			t.Fatalf("register %s: %d %s", username, created.Code, created.Body.String())
		}
		login := performJSON(t, target, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
			"username": username, "password": "Password1!",
		})
		if login.Code != http.StatusOK {
			t.Fatalf("login %s: %d %s", username, login.Code, login.Body.String())
		}
		var body struct {
			Token string `json:"token"`
		}
		decode(t, login.Body, &body)
		return body.Token
	}
	ownerToken := registerAndLogin("reviewowner")
	otherToken := registerAndLogin("reviewother")
	assertAPIError(t, performJSON(t, target, http.MethodGet, "/api/v1/reviews?cursor=bad", "", nil),
		http.StatusBadRequest, "INVALID_CURSOR")
	assertAPIError(t, performJSON(t, target, http.MethodGet, "/api/v1/reviews/0", "", nil),
		http.StatusBadRequest, "INVALID_REVIEW_ID")
	assertAPIError(t, performJSON(t, target, http.MethodGet, "/api/v1/reviews/me", ownerToken, nil),
		http.StatusNotFound, "REVIEW_NOT_FOUND")
	assertAPIError(t, performJSON(t, target, http.MethodPost, "/api/v1/reviews", "", map[string]any{
		"body": "Nice", "rating": 5,
	}), http.StatusUnauthorized, "UNAUTHORIZED")
	for _, invalid := range []map[string]any{
		{"body": " ", "rating": 5}, {"body": "Nice", "rating": 0},
		{"body": "Nice"}, {"body": "Nice", "rating": 6},
		{"body": strings.Repeat("a", 2001), "rating": 5},
	} {
		assertAPIError(t, performJSON(t, target, http.MethodPost, "/api/v1/reviews", ownerToken, invalid),
			http.StatusUnprocessableEntity, "INVALID_FORM_BODY")
	}
	created := performJSON(t, target, http.MethodPost, "/api/v1/reviews", ownerToken, map[string]any{
		"body": "Nice", "rating": 5,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create review: %d %s", created.Code, created.Body.String())
	}
	var root struct {
		ID string `json:"id"`
	}
	decode(t, created.Body, &root)
	if root.ID == "" {
		t.Fatal("root ID missing")
	}
	assertAPIError(t, performJSON(t, target, http.MethodPost, "/api/v1/reviews", ownerToken, map[string]any{
		"body": "Again", "rating": 4,
	}), http.StatusConflict, "REVIEW_ALREADY_EXISTS")
	rootPath := "/api/v1/reviews/" + root.ID
	assertAPIError(t, performJSON(t, target, http.MethodPost, "/api/v1/reviews/999999999/replies", otherToken,
		map[string]any{"body": "Orphan"}), http.StatusNotFound, "REVIEW_NOT_FOUND")
	assertAPIError(t, performJSON(t, target, http.MethodGet, rootPath+"/replies", "", nil),
		http.StatusUnauthorized, "UNAUTHORIZED")
	assertAPIError(t, performJSON(t, target, http.MethodPost, rootPath+"/replies", otherToken, map[string]any{
		"body": "No rating", "rating": nil,
	}), http.StatusUnprocessableEntity, "INVALID_FORM_BODY")
	replied := performJSON(t, target, http.MethodPost, rootPath+"/replies", otherToken, map[string]any{"body": "Agree"})
	if replied.Code != http.StatusCreated {
		t.Fatalf("create reply: %d %s", replied.Code, replied.Body.String())
	}
	var reply struct {
		ID string `json:"id"`
	}
	decode(t, replied.Body, &reply)
	deep := performJSON(t, target, http.MethodPost, "/api/v1/reviews/"+reply.ID+"/replies", ownerToken, map[string]any{
		"body": "Thanks",
	})
	if deep.Code != http.StatusCreated {
		t.Fatalf("create nested reply: %d %s", deep.Code, deep.Body.String())
	}
	for _, path := range []string{"/api/v1/reviews", rootPath, "/api/v1/reviews/me"} {
		token := ownerToken
		if path == "/api/v1/reviews" || path == rootPath {
			token = ""
		}
		result := performJSON(t, target, http.MethodGet, path, token, nil)
		if result.Code != http.StatusOK {
			t.Fatalf("read %s: %d %s", path, result.Code, result.Body.String())
		}
	}
	publicList := performJSON(t, target, http.MethodGet, "/api/v1/reviews", "", nil)
	var publicBody struct {
		Items []map[string]any `json:"items"`
	}
	decode(t, publicList.Body, &publicBody)
	if len(publicBody.Items) != 1 {
		t.Fatalf("public roots: %d", len(publicBody.Items))
	}
	if _, present := publicBody.Items[0]["replies"]; present {
		t.Fatal("public roots included replies")
	}
	invalidTokenRequest := performJSON(t, target, http.MethodGet, "/api/v1/reviews", "not-a-real-token", nil)
	assertAPIError(t, invalidTokenRequest, http.StatusUnauthorized, "UNAUTHORIZED")
	authList := performJSON(t, target, http.MethodGet, "/api/v1/reviews", ownerToken, nil)
	var authenticatedBody struct {
		Items []struct {
			Replies struct {
				Items []struct {
					ID string `json:"id"`
				} `json:"items"`
			} `json:"replies"`
		} `json:"items"`
	}
	decode(t, authList.Body, &authenticatedBody)
	if len(authenticatedBody.Items) != 1 || len(authenticatedBody.Items[0].Replies.Items) != 1 ||
		authenticatedBody.Items[0].Replies.Items[0].ID != reply.ID {
		t.Fatalf("authenticated preview missing: %+v", authenticatedBody)
	}
	assertAPIError(t, performJSON(t, target, http.MethodGet, "/api/v1/reviews/"+reply.ID, "", nil),
		http.StatusUnauthorized, "UNAUTHORIZED")
	children := performJSON(t, target, http.MethodGet, rootPath+"/replies", ownerToken, nil)
	var childrenBody struct {
		Items []map[string]any `json:"items"`
	}
	decode(t, children.Body, &childrenBody)
	if len(childrenBody.Items) != 1 || childrenBody.Items[0]["has_replies"] != true {
		t.Fatalf("direct replies or child flag missing: %+v", childrenBody)
	}
	if _, present := childrenBody.Items[0]["replies"]; present {
		t.Fatal("reply list included nested content")
	}
	assertAPIError(t, performJSON(t, target, http.MethodDelete, "/api/v1/reviews/"+reply.ID, ownerToken, nil),
		http.StatusNotFound, "REVIEW_NOT_FOUND")
	deletedReply := performJSON(t, target, http.MethodDelete, "/api/v1/reviews/"+reply.ID, otherToken, nil)
	if deletedReply.Code != http.StatusNoContent {
		t.Fatalf("delete reply: %d %s", deletedReply.Code, deletedReply.Body.String())
	}
	tombstone := performJSON(t, target, http.MethodGet, "/api/v1/reviews/"+reply.ID, ownerToken, nil)
	var deletedBody map[string]any
	decode(t, tombstone.Body, &deletedBody)
	if deletedBody["body"] != nil || deletedBody["author"] != nil || deletedBody["deleted"] != true {
		t.Fatalf("tombstone leaked content: %+v", deletedBody)
	}
	if result := performJSON(t, target, http.MethodDelete, rootPath, ownerToken, nil); result.Code != http.StatusNoContent {
		t.Fatalf("delete root: %d %s", result.Code, result.Body.String())
	}
	assertAPIError(t, performJSON(t, target, http.MethodGet, rootPath, "", nil),
		http.StatusNotFound, "REVIEW_NOT_FOUND")
	assertAPIError(t, performJSON(t, target, http.MethodPost, rootPath+"/replies", otherToken,
		map[string]any{"body": "Too late"}), http.StatusNotFound, "REVIEW_NOT_FOUND")
	assertAPIError(t, performJSON(t, target, http.MethodGet, "/api/v1/reviews/"+reply.ID, ownerToken, nil),
		http.StatusNotFound, "REVIEW_NOT_FOUND")
	assertAPIError(t, performJSON(t, target, http.MethodGet, "/api/v1/reviews/me", ownerToken, nil),
		http.StatusNotFound, "REVIEW_NOT_FOUND")
	newRoot := performJSON(t, target, http.MethodPost, "/api/v1/reviews", ownerToken, map[string]any{
		"body": "New review", "rating": 4,
	})
	if newRoot.Code != http.StatusCreated {
		t.Fatalf("repost review: %d %s", newRoot.Code, newRoot.Body.String())
	}
}
