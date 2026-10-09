package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
	"github.com/DulioShortener/Shortener/backend/internal/service"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session entity.AuthSession) (entity.AuthSession, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO auth_sessions (id, token_hash, user_id, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`, session.ID, session.TokenHash, session.UserID,
		unixMilliseconds(session.CreatedAt), unixMilliseconds(session.ExpiresAt))
	if err != nil {
		return entity.AuthSession{}, fmt.Errorf("insert authentication session: %w", err)
	}
	return session, nil
}

func (r *SessionRepository) FindByTokenHash(ctx context.Context, tokenHash []byte) (entity.AuthSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, token_hash, user_id, created_at, expires_at
		FROM auth_sessions
		WHERE token_hash = ?
	`, tokenHash)

	var session entity.AuthSession
	var createdAt, expiresAt int64
	if err := row.Scan(
		&session.ID, &session.TokenHash, &session.UserID, &createdAt, &expiresAt,
	); errors.Is(err, sql.ErrNoRows) {
		return entity.AuthSession{}, service.ErrUnauthorized
	} else if err != nil {
		return entity.AuthSession{}, fmt.Errorf("query authentication session: %w", err)
	}
	session.CreatedAt = timeFromMilliseconds(createdAt)
	session.ExpiresAt = timeFromMilliseconds(expiresAt)
	return session, nil
}

func (r *SessionRepository) Delete(ctx context.Context, id, userID int64) error {
	if _, err := r.db.ExecContext(ctx, `
		DELETE FROM auth_sessions WHERE id = ? AND user_id = ?
	`, id, userID); err != nil {
		return fmt.Errorf("delete authentication session: %w", err)
	}
	return nil
}
