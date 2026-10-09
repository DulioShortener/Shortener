package entity

import "time"

type AuthSession struct {
	ID        int64
	TokenHash []byte
	UserID    int64
	CreatedAt time.Time
	ExpiresAt time.Time
}
