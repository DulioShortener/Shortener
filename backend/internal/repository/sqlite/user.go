package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
	"github.com/DulioShortener/Shortener/backend/internal/service"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user entity.User) (entity.User, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, username, display_name, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, user.ID, user.Username, user.DisplayName, user.PasswordHash,
		unixMilliseconds(user.CreatedAt), unixMilliseconds(user.UpdatedAt))
	if constraintContains(err, "users.username") {
		return entity.User{}, service.ErrUsernameTaken
	}
	if err != nil {
		return entity.User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (entity.User, error) {
	return r.find(ctx, "id = ?", id)
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (entity.User, error) {
	return r.find(ctx, "username = ?", username)
}

func (r *UserRepository) find(ctx context.Context, predicate string, argument any) (entity.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, display_name, password_hash, created_at, updated_at
		FROM users
		WHERE `+predicate, argument)

	var user entity.User
	var displayName sql.NullString
	var createdAt, updatedAt int64
	if err := row.Scan(
		&user.ID, &user.Username, &displayName, &user.PasswordHash, &createdAt, &updatedAt,
	); errors.Is(err, sql.ErrNoRows) {
		return entity.User{}, service.ErrUserNotFound
	} else if err != nil {
		return entity.User{}, fmt.Errorf("query user: %w", err)
	}
	if displayName.Valid {
		user.DisplayName = &displayName.String
	}
	user.CreatedAt = timeFromMilliseconds(createdAt)
	user.UpdatedAt = timeFromMilliseconds(updatedAt)
	return user, nil
}
