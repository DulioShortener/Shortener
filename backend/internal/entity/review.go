package entity

import "time"

// ReviewComment is a rated root review or an unrated reply.
type ReviewComment struct {
	ID          int64
	UserID      int64
	ParentID    *int64
	Body        string
	Rating      *int
	CreatedAt   time.Time
	DeletedAt   *time.Time
	Username    string
	DisplayName *string
	HasReplies  bool
}

type ReviewCursor struct {
	CreatedAt time.Time
	ID        int64
}

type ReviewPage struct {
	Items      []ReviewComment
	HasMore    bool
	NextCursor *ReviewCursor
}
