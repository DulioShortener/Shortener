package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
	"github.com/DulioShortener/Shortener/backend/internal/service"
)

type LinkRepository struct {
	db *sql.DB
}

func NewLinkRepository(db *sql.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) Create(ctx context.Context, link entity.Link) (entity.Link, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO links (id, user_id, code, target_url, target_url_key, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, link.ID, link.UserID, link.Code, link.TargetURL, link.TargetURLKey,
		unixMilliseconds(link.CreatedAt))
	if constraintContains(err, "LINK_LIMIT_REACHED") {
		return entity.Link{}, service.ErrLinkLimitReached
	}
	if constraintContains(err, "links.user_id, links.target_url_key") {
		return entity.Link{}, service.ErrLinkAlreadyExists
	}
	if constraintContains(err, "links.code") {
		return entity.Link{}, service.ErrShortCodeCollision
	}
	if err != nil {
		return entity.Link{}, fmt.Errorf("insert link: %w", err)
	}
	return link, nil
}

func (r *LinkRepository) ListByUser(ctx context.Context, userID int64) ([]entity.Link, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, code, target_url, target_url_key, created_at
		FROM links
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}
	defer rows.Close()

	links := make([]entity.Link, 0)
	for rows.Next() {
		link, err := scanLink(rows)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate links: %w", err)
	}
	return links, nil
}

func (r *LinkRepository) Delete(ctx context.Context, id, userID int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM links WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted link count: %w", err)
	}
	if affected == 0 {
		return service.ErrLinkNotFound
	}
	return nil
}

func (r *LinkRepository) FindByCode(ctx context.Context, code string) (entity.Link, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, code, target_url, target_url_key, created_at
		FROM links
		WHERE code = ?
	`, code)
	link, err := scanLink(row)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Link{}, service.ErrLinkNotFound
	}
	return link, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanLink(row scanner) (entity.Link, error) {
	var link entity.Link
	var createdAt int64
	if err := row.Scan(
		&link.ID, &link.UserID, &link.Code, &link.TargetURL, &link.TargetURLKey, &createdAt,
	); err != nil {
		return entity.Link{}, err
	}
	link.CreatedAt = timeFromMilliseconds(createdAt)
	return link, nil
}
