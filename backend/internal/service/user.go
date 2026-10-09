package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user entity.User) (entity.User, error)
	FindByID(ctx context.Context, id int64) (entity.User, error)
	FindByUsername(ctx context.Context, username string) (entity.User, error)
}

type UserCreateInput struct {
	Username    string
	DisplayName *string
	Password    string
}

type UserService struct {
	users     UserRepository
	passwords PasswordHasher
	ids       IDGenerator
	clock     Clock
}

func NewUserService(users UserRepository, passwords PasswordHasher, ids IDGenerator, clock Clock) *UserService {
	return &UserService{users: users, passwords: passwords, ids: ids, clock: clock}
}

func (s *UserService) Create(ctx context.Context, input UserCreateInput) (entity.User, error) {
	passwordHash, err := s.passwords.Hash(input.Password)
	if err != nil {
		return entity.User{}, fmt.Errorf("hash password: %w", err)
	}
	id, err := s.ids.NextID()
	if err != nil {
		return entity.User{}, err
	}

	var displayName *string
	if input.DisplayName != nil {
		value := strings.TrimSpace(*input.DisplayName)
		displayName = &value
	}
	now := s.clock.Now().UTC()
	return s.users.Create(ctx, entity.User{
		ID:           id,
		Username:     strings.ToLower(input.Username),
		DisplayName:  displayName,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func (s *UserService) Get(ctx context.Context, id int64) (entity.User, error) {
	return s.users.FindByID(ctx, id)
}
