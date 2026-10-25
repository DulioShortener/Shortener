package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DulioShortener/Shortener/backend/internal/apierrors"
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
	registerAndLogin := func(username string) (string, string) {
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
			User  struct {
				ID string `json:"id"`
			} `json:"user"`
		}
		decode(t, login.Body, &body)
		return body.Token, body.User.ID
	}
	ownerToken, ownerID := registerAndLogin("reviewowner")
	otherToken, otherID := registerAndLogin("reviewother")
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
		result := performJSON(t, target, http.MethodPost, "/api/v1/reviews", ownerToken, invalid)
		if rating, ok := invalid["rating"].(int); ok && (rating < 1 || rating > 5) {
			var body apierrors.Response
			decode(t, bytes.NewReader(result.Body.Bytes()), &body)
			if body.Errors.Fields["rating"] != "This value must be in range of 1 and 5" {
				t.Fatalf("unexpected rating error: %+v", body.Errors.Fields)
			}
		}
		assertAPIError(t, result, http.StatusUnprocessableEntity, "INVALID_FORM_BODY")
	}
	created := performJSON(t, target, http.MethodPost, "/api/v1/reviews", ownerToken, map[string]any{
		"body": "Nice", "rating": 5,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create review: %d %s", created.Code, created.Body.String())
	}
	var root struct {
		ID       string `json:"id"`
		AuthorID string `json:"author_id"`
	}
	decode(t, bytes.NewReader(created.Body.Bytes()), &root)
	if root.ID == "" {
		t.Fatal("root ID missing")
	}
	if root.AuthorID != ownerID {
		t.Fatalf("root author_id=%q, want %q", root.AuthorID, ownerID)
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
		ID       string `json:"id"`
		AuthorID string `json:"author_id"`
	}
	decode(t, bytes.NewReader(replied.Body.Bytes()), &reply)
	if reply.AuthorID != otherID {
		t.Fatalf("reply author_id=%q, want %q", reply.AuthorID, otherID)
	}
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
		assertNoReviewMetadata(t, result.Body.Bytes())
	}
	publicList := performJSON(t, target, http.MethodGet, "/api/v1/reviews", "", nil)
	var publicBody struct {
		Items []map[string]any `json:"items"`
	}
	decode(t, bytes.NewReader(publicList.Body.Bytes()), &publicBody)
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
	decode(t, bytes.NewReader(authList.Body.Bytes()), &authenticatedBody)
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
	decode(t, bytes.NewReader(children.Body.Bytes()), &childrenBody)
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
	decode(t, bytes.NewReader(tombstone.Body.Bytes()), &deletedBody)
	if body, present := deletedBody["body"]; !present || body != nil {
		t.Fatalf("tombstone leaked content: %+v", deletedBody)
	}
	if deletedBody["author_id"] != otherID {
		t.Fatalf("reply placeholder lost author_id: %+v", deletedBody)
	}
	for _, result := range [][]byte{
		created.Body.Bytes(), replied.Body.Bytes(), deep.Body.Bytes(), publicList.Body.Bytes(),
		authList.Body.Bytes(), children.Body.Bytes(), tombstone.Body.Bytes(),
	} {
		assertNoReviewMetadata(t, result)
	}
	for _, path := range []string{
		"/api/v1/reviews", "/api/v1/reviews/me", rootPath,
		rootPath + "/replies", "/api/v1/reviews/" + reply.ID + "/replies",
	} {
		result := performJSON(t, target, http.MethodGet, path, ownerToken, nil)
		if result.Code != http.StatusOK {
			t.Fatalf("read after deleting reply %s: %d %s", path, result.Code, result.Body.String())
		}
		assertNoReviewMetadata(t, result.Body.Bytes())
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
	assertNoReviewMetadata(t, newRoot.Body.Bytes())
}

func assertNoReviewMetadata(t *testing.T, body []byte) {
	t.Helper()
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		t.Fatal(err)
	}
	var check func(any)
	check = func(value any) {
		switch value := value.(type) {
		case map[string]any:
			if _, comment := value["body"]; comment {
				authorID, ok := value["author_id"].(string)
				id, err := strconv.ParseInt(authorID, 10, 64)
				if !ok || err != nil || id <= 0 || strconv.FormatInt(id, 10) != authorID {
					t.Errorf("review response has invalid author_id: %s", body)
				}
			}
			for key, child := range value {
				switch key {
				case "user", "author", "user_id", "username", "display_name", "delete", "deleted", "deleted_at":
					t.Errorf("review response exposed %q: %s", key, body)
				}
				check(child)
			}
		case []any:
			for _, child := range value {
				check(child)
			}
		}
	}
	check(value)
}
