package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/DulioShortener/Shortener/backend/internal/service"
)

const tokenBytes = 32

type Generator struct {
	random io.Reader
}

func NewGenerator() *Generator {
	return &Generator{random: rand.Reader}
}

func (g *Generator) Generate() (service.GeneratedToken, error) {
	value := make([]byte, tokenBytes)
	if _, err := io.ReadFull(g.random, value); err != nil {
		return service.GeneratedToken{}, fmt.Errorf("generate authentication token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(value)
	return service.GeneratedToken{Raw: raw, Hash: g.Hash(raw)}, nil
}

func (g *Generator) Hash(raw string) []byte {
	hash := sha256.Sum256([]byte(raw))
	return hash[:]
}
