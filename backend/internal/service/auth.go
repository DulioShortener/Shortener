package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
)

type SessionRepository interface {
	Create(ctx context.Context, session entity.AuthSession) (entity.AuthSession, error)
	FindByTokenHash(ctx context.Context, tokenHash []byte) (entity.AuthSession, error)
	Delete(ctx context.Context, id, userID int64) error
}

type LoginInput struct {
	Username string
	Password string
}

type LoginResult struct {
	Session entity.AuthSession
	User    entity.User
	Token   string
}

type Principal struct {
	UserID    int64
	SessionID int64
}

type AuthService struct {
	users         UserRepository
	sessions      SessionRepository
	passwords     PasswordHasher
	tokens        TokenGenerator
	ids           IDGenerator
	clock         Clock
	tokenTTL      time.Duration
	dummyPassword string
}

func NewAuthService(
	users UserRepository,
	sessions SessionRepository,
	passwords PasswordHasher,
	tokens TokenGenerator,
	ids IDGenerator,
	clock Clock,
	tokenTTL time.Duration,
) (*AuthService, error) {
	dummyPassword, err := passwords.Hash("DulioDummyPassword1!")
	if err != nil {
		return nil, fmt.Errorf("create dummy password hash: %w", err)
	}
	return &AuthService{
		users: users, sessions: sessions, passwords: passwords, tokens: tokens,
		ids: ids, clock: clock, tokenTTL: tokenTTL, dummyPassword: dummyPassword,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	user, err := s.users.FindByUsername(ctx, strings.ToLower(input.Username))
	if errors.Is(err, ErrUserNotFound) {
		_, _ = s.passwords.Verify(input.Password, s.dummyPassword)
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, err
	}

	valid, err := s.passwords.Verify(input.Password, user.PasswordHash)
	if err != nil {
		return LoginResult{}, fmt.Errorf("verify password: %w", err)
	}
	if !valid {
		return LoginResult{}, ErrInvalidCredentials
	}

	generatedToken, err := s.tokens.Generate()
	if err != nil {
		return LoginResult{}, err
	}
	id, err := s.ids.NextID()
	if err != nil {
		return LoginResult{}, err
	}
	now := s.clock.Now().UTC()
	session, err := s.sessions.Create(ctx, entity.AuthSession{
		ID: id, TokenHash: generatedToken.Hash, UserID: user.ID,
		CreatedAt: now, ExpiresAt: now.Add(s.tokenTTL),
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Session: session, User: user, Token: generatedToken.Raw}, nil
}

func (s *AuthService) Authenticate(ctx context.Context, rawToken string) (Principal, error) {
	if rawToken == "" {
		return Principal{}, ErrUnauthorized
	}
	session, err := s.sessions.FindByTokenHash(ctx, s.tokens.Hash(rawToken))
	if errors.Is(err, ErrUnauthorized) {
		return Principal{}, ErrUnauthorized
	}
	if err != nil {
		return Principal{}, err
	}
	if !session.ExpiresAt.After(s.clock.Now()) {
		_ = s.sessions.Delete(ctx, session.ID, session.UserID)
		return Principal{}, ErrUnauthorized
	}
	return Principal{UserID: session.UserID, SessionID: session.ID}, nil
}

func (s *AuthService) Logout(ctx context.Context, principal Principal) error {
	if err := s.sessions.Delete(ctx, principal.SessionID, principal.UserID); err != nil && !errors.Is(err, ErrUnauthorized) {
		return err
	}
	return nil
}
