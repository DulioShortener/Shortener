package service

import "errors"

var (
	ErrUsernameTaken       = errors.New("username is already taken")
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrUnauthorized        = errors.New("authentication is required")
	ErrUserNotFound        = errors.New("user not found")
	ErrLinkAlreadyExists   = errors.New("link already exists for user")
	ErrLinkLimitReached    = errors.New("link limit reached")
	ErrLinkNotFound        = errors.New("link not found")
	ErrShortCodeCollision  = errors.New("short code collision")
	ErrShortCodeExhausted  = errors.New("could not allocate a unique short code")
	ErrReviewNotFound      = errors.New("review comment not found")
	ErrReviewAlreadyExists = errors.New("active review already exists")
	ErrInvalidReview       = errors.New("invalid review comment")
)
