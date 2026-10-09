package service

import "time"

type IDGenerator interface {
	NextID() (int64, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encodedHash string) (bool, error)
}

type GeneratedToken struct {
	Raw  string
	Hash []byte
}

type TokenGenerator interface {
	Generate() (GeneratedToken, error)
	Hash(raw string) []byte
}

type CodeGenerator interface {
	Generate() (string, error)
}

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}
