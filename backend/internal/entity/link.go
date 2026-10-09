package entity

import "time"

type Link struct {
	ID           int64
	UserID       int64
	Code         string
	TargetURL    string
	TargetURLKey string
	CreatedAt    time.Time
}
