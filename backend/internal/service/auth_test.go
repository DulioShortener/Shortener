package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
)

type authUserRepository struct {
	user entity.User
	err  error
}

func (r *authUserRepository) Create(context.Context, entity.User) (entity.User, error) {
	panic("not used")
}

func (r *authUserRepository) FindByID(context.Context, int64) (entity.User, error) {
	panic("not used")
}

func (r *authUserRepository) FindByUsername(context.Context, string) (entity.User, error) {
	return r.user, r.err
}

type authSessionRepository struct {
	session entity.AuthSession
	deleted bool
}

func (r *authSessionRepository) Create(_ context.Context, session entity.AuthSession) (entity.AuthSession, error) {
	return session, nil
}

func (r *authSessionRepository) FindByTokenHash(context.Context, []byte) (entity.AuthSession, error) {
	return r.session, nil
}

func (r *authSessionRepository) Delete(context.Context, int64, int64) error {
	r.deleted = true
	return nil
}

type authPasswordHasher struct {
	verifyCalls int
}

func (h *authPasswordHasher) Hash(string) (string, error) { return "dummy-hash", nil }

func (h *authPasswordHasher) Verify(string, string) (bool, error) {
	h.verifyCalls++
	return false, nil
}

type authTokenGenerator struct{}

func (authTokenGenerator) Generate() (GeneratedToken, error) {
	return GeneratedToken{Raw: "token", Hash: make([]byte, 32)}, nil
}

func (authTokenGenerator) Hash(string) []byte { return make([]byte, 32) }

type fixedID int64

func (id fixedID) NextID() (int64, error) { return int64(id), nil }

type fixedClock time.Time

func (clock fixedClock) Now() time.Time { return time.Time(clock) }

func TestLoginUsesDummyHashForUnknownUsername(t *testing.T) {
	hasher := &authPasswordHasher{}
	auth, err := NewAuthService(
		&authUserRepository{err: ErrUserNotFound}, &authSessionRepository{}, hasher,
		authTokenGenerator{}, fixedID(1), fixedClock(time.Now()), time.Hour,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = auth.Login(context.Background(), LoginInput{Username: "missing", Password: "guess"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if hasher.verifyCalls != 1 {
		t.Fatalf("expected dummy hash verification, got %d calls", hasher.verifyCalls)
	}
}

func TestAuthenticateRejectsAndDeletesExpiredSession(t *testing.T) {
	now := time.Now().UTC()
	sessions := &authSessionRepository{session: entity.AuthSession{
		ID: 10, UserID: 20, ExpiresAt: now.Add(-time.Second),
	}}
	auth, err := NewAuthService(
		&authUserRepository{}, sessions, &authPasswordHasher{}, authTokenGenerator{},
		fixedID(1), fixedClock(now), time.Hour,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = auth.Authenticate(context.Background(), "expired")
	if !errors.Is(err, ErrUnauthorized) || !sessions.deleted {
		t.Fatalf("expected expired session rejection and deletion, err=%v deleted=%v", err, sessions.deleted)
	}
}
