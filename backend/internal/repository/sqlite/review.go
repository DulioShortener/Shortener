package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
	"github.com/DulioShortener/Shortener/backend/internal/service"
)

const (
	reviewPageLimit = 50
	reviewSelect    = `c.id, c.user_id, c.parent_id, c.body, c.rating, c.created_at,
		c.deleted_at, u.username, u.display_name,
		EXISTS(SELECT 1 FROM review_comments child WHERE child.parent_id = c.id)`
	reviewJoin     = `FROM review_comments c JOIN users u ON u.id = c.user_id`
	reviewAncestry = `WITH RECURSIVE ancestry(id, parent_id, deleted_at) AS (
		SELECT id, parent_id, deleted_at FROM review_comments WHERE id = ?
		UNION ALL
		SELECT parent.id, parent.parent_id, parent.deleted_at
		FROM review_comments parent JOIN ancestry child ON parent.id = child.parent_id
	)`
	reviewVisible = `EXISTS(SELECT 1 FROM ancestry WHERE parent_id IS NULL AND deleted_at IS NULL)`
)

type ReviewRepository struct {
	db *sql.DB
}

func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) CreateRoot(ctx context.Context, comment entity.ReviewComment) (entity.ReviewComment, error) {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO review_comments(id, user_id, body, rating, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, comment.ID, comment.UserID, comment.Body, *comment.Rating, unixMilliseconds(comment.CreatedAt))
	if constraintContains(err, "review_comments.user_id") {
		return entity.ReviewComment{}, service.ErrReviewAlreadyExists
	}
	if err != nil {
		return entity.ReviewComment{}, fmt.Errorf("insert review: %w", err)
	}
	return r.FindVisible(ctx, comment.ID)
}

func (r *ReviewRepository) CreateReply(ctx context.Context, comment entity.ReviewComment) (entity.ReviewComment, error) {
	result, err := r.db.ExecContext(ctx, reviewAncestry+`
		INSERT INTO review_comments(id, user_id, parent_id, body, created_at)
		SELECT ?, ?, ?, ?, ? WHERE `+reviewVisible+`
		AND EXISTS(SELECT 1 FROM ancestry WHERE id = ? AND deleted_at IS NULL)`,
		*comment.ParentID, comment.ID, comment.UserID, *comment.ParentID,
		comment.Body, unixMilliseconds(comment.CreatedAt), *comment.ParentID)
	if err != nil {
		return entity.ReviewComment{}, fmt.Errorf("insert reply: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return entity.ReviewComment{}, fmt.Errorf("read inserted reply count: %w", err)
	}
	if affected == 0 {
		return entity.ReviewComment{}, service.ErrReviewNotFound
	}
	return r.FindVisible(ctx, comment.ID)
}

func (r *ReviewRepository) FindVisible(ctx context.Context, id int64) (entity.ReviewComment, error) {
	row := r.db.QueryRowContext(ctx, reviewAncestry+`
		SELECT `+reviewSelect+` `+reviewJoin+`
		WHERE c.id = ? AND `+reviewVisible,
		id, id)
	comment, err := scanReview(row)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.ReviewComment{}, service.ErrReviewNotFound
	}
	if err != nil {
		return entity.ReviewComment{}, fmt.Errorf("find review comment: %w", err)
	}
	return comment, nil
}

func (r *ReviewRepository) FindActiveRootByUser(ctx context.Context, userID int64) (entity.ReviewComment, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+reviewSelect+` `+reviewJoin+`
		WHERE c.user_id = ? AND c.parent_id IS NULL AND c.deleted_at IS NULL`, userID)
	comment, err := scanReview(row)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.ReviewComment{}, service.ErrReviewNotFound
	}
	if err != nil {
		return entity.ReviewComment{}, fmt.Errorf("find own review: %w", err)
	}
	return comment, nil
}

func (r *ReviewRepository) ListRoots(ctx context.Context, cursor *entity.ReviewCursor, includePreviews bool) (service.ReviewListing, error) {
	var result service.ReviewListing
	result.Previews = make(map[int64]entity.ReviewPage)
	transaction, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return result, fmt.Errorf("begin review list: %w", err)
	}
	defer transaction.Rollback()
	query := `SELECT ` + reviewSelect + ` ` + reviewJoin + `
		WHERE c.parent_id IS NULL AND c.deleted_at IS NULL`
	arguments := make([]any, 0, 3)
	if cursor != nil {
		query += ` AND (c.created_at < ? OR (c.created_at = ? AND c.id < ?))`
		arguments = append(arguments, unixMilliseconds(cursor.CreatedAt), unixMilliseconds(cursor.CreatedAt), cursor.ID)
	}
	query += ` ORDER BY c.created_at DESC, c.id DESC LIMIT 51`
	rows, err := transaction.QueryContext(ctx, query, arguments...)
	if err != nil {
		return result, fmt.Errorf("list reviews: %w", err)
	}
	result.Page, err = scanReviewPage(rows, reviewPageLimit)
	if err != nil {
		return result, err
	}
	if includePreviews && len(result.Page.Items) > 0 {
		result.Previews, err = r.rootPreviews(ctx, transaction, result.Page.Items)
		if err != nil {
			return result, err
		}
	}
	if err := transaction.Commit(); err != nil {
		return result, fmt.Errorf("commit review list: %w", err)
	}
	return result, nil
}

func (r *ReviewRepository) rootPreviews(ctx context.Context, transaction *sql.Tx, roots []entity.ReviewComment) (map[int64]entity.ReviewPage, error) {
	previews := make(map[int64]entity.ReviewPage, len(roots))
	placeholders := make([]string, 0, len(roots))
	arguments := make([]any, 0, len(roots))
	for _, root := range roots {
		previews[root.ID] = entity.ReviewPage{Items: []entity.ReviewComment{}}
		placeholders = append(placeholders, "?")
		arguments = append(arguments, root.ID)
	}
	query := `SELECT ` + reviewSelect + `,
		EXISTS(SELECT 1 FROM review_comments later
			WHERE later.parent_id = c.parent_id
			ORDER BY later.created_at, later.id LIMIT 1 OFFSET 1)
		` + reviewJoin + `
		WHERE c.parent_id IN (` + strings.Join(placeholders, ",") + `)
		AND c.id = (SELECT first.id FROM review_comments first
			WHERE first.parent_id = c.parent_id
			ORDER BY first.created_at, first.id LIMIT 1)`
	rows, err := transaction.QueryContext(ctx, query, arguments...)
	if err != nil {
		return nil, fmt.Errorf("list root reply previews: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var later bool
		comment, err := scanReviewWithExtra(rows, &later)
		if err != nil {
			return nil, fmt.Errorf("scan root reply preview: %w", err)
		}
		preview := entity.ReviewPage{Items: []entity.ReviewComment{comment}, HasMore: later}
		if later {
			preview.NextCursor = cursorAfter(comment)
		}
		previews[*comment.ParentID] = preview
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate root reply previews: %w", err)
	}
	return previews, nil
}

func (r *ReviewRepository) ListReplies(ctx context.Context, parentID int64, cursor *entity.ReviewCursor) (entity.ReviewPage, error) {
	if _, err := r.FindVisible(ctx, parentID); err != nil {
		return entity.ReviewPage{}, err
	}
	return r.listChildren(ctx, parentID, cursor, reviewPageLimit)
}

func (r *ReviewRepository) Preview(ctx context.Context, parentID int64) (entity.ReviewPage, error) {
	return r.listChildren(ctx, parentID, nil, 1)
}

func (r *ReviewRepository) listChildren(ctx context.Context, parentID int64, cursor *entity.ReviewCursor, limit int) (entity.ReviewPage, error) {
	query := reviewAncestry + ` SELECT ` + reviewSelect + ` ` + reviewJoin + `
		WHERE c.parent_id = ? AND ` + reviewVisible
	arguments := []any{parentID, parentID}
	if cursor != nil {
		query += ` AND (c.created_at > ? OR (c.created_at = ? AND c.id > ?))`
		arguments = append(arguments, unixMilliseconds(cursor.CreatedAt), unixMilliseconds(cursor.CreatedAt), cursor.ID)
	}
	query += ` ORDER BY c.created_at, c.id LIMIT ?`
	arguments = append(arguments, limit+1)
	rows, err := r.db.QueryContext(ctx, query, arguments...)
	if err != nil {
		return entity.ReviewPage{}, fmt.Errorf("list replies: %w", err)
	}
	return scanReviewPage(rows, limit)
}

func (r *ReviewRepository) SoftDelete(ctx context.Context, id, userID int64, deletedAt time.Time) error {
	result, err := r.db.ExecContext(ctx, reviewAncestry+`
		UPDATE review_comments SET deleted_at = ?
		WHERE id = ? AND user_id = ? AND deleted_at IS NULL AND `+reviewVisible,
		id, unixMilliseconds(deletedAt), id, userID)
	if err != nil {
		return fmt.Errorf("soft-delete review comment: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted review count: %w", err)
	}
	if affected == 0 {
		return service.ErrReviewNotFound
	}
	return nil
}

func scanReviewPage(rows *sql.Rows, limit int) (entity.ReviewPage, error) {
	defer rows.Close()
	page := entity.ReviewPage{Items: make([]entity.ReviewComment, 0, limit)}
	for rows.Next() {
		comment, err := scanReview(rows)
		if err != nil {
			return entity.ReviewPage{}, fmt.Errorf("scan review comment: %w", err)
		}
		if len(page.Items) == limit {
			page.HasMore = true
			break
		}
		page.Items = append(page.Items, comment)
	}
	if err := rows.Err(); err != nil {
		return entity.ReviewPage{}, fmt.Errorf("iterate review comments: %w", err)
	}
	if page.HasMore {
		page.NextCursor = cursorAfter(page.Items[len(page.Items)-1])
	}
	return page, nil
}

func cursorAfter(comment entity.ReviewComment) *entity.ReviewCursor {
	return &entity.ReviewCursor{CreatedAt: comment.CreatedAt, ID: comment.ID}
}

func scanReview(row scanner) (entity.ReviewComment, error) {
	return scanReviewWithExtra(row)
}

func scanReviewWithExtra(row scanner, extra ...any) (entity.ReviewComment, error) {
	var comment entity.ReviewComment
	var parentID, rating, deletedAt sql.NullInt64
	var displayName sql.NullString
	var createdAt int64
	destinations := []any{
		&comment.ID, &comment.UserID, &parentID, &comment.Body, &rating, &createdAt,
		&deletedAt, &comment.Username, &displayName, &comment.HasReplies,
	}
	destinations = append(destinations, extra...)
	if err := row.Scan(destinations...); err != nil {
		return entity.ReviewComment{}, err
	}
	if parentID.Valid {
		comment.ParentID = &parentID.Int64
	}
	if rating.Valid {
		value := int(rating.Int64)
		comment.Rating = &value
	}
	if deletedAt.Valid {
		value := timeFromMilliseconds(deletedAt.Int64)
		comment.DeletedAt = &value
	}
	if displayName.Valid {
		comment.DisplayName = &displayName.String
	}
	comment.CreatedAt = timeFromMilliseconds(createdAt)
	return comment, nil
}
