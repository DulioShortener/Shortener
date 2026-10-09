package entity

import "time"

type User struct {
	ID           int64
	Username     string
	DisplayName  *string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
