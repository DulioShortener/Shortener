package shortcode

import (
	"crypto/rand"
	"fmt"
	"io"
)

const (
	alphabet   = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	codeLength = 8
	maxByte    = 256 - (256 % len(alphabet))
)

type Generator struct {
	random io.Reader
}

func NewGenerator() *Generator {
	return &Generator{random: rand.Reader}
}

func (g *Generator) Generate() (string, error) {
	code := make([]byte, 0, codeLength)
	buffer := make([]byte, codeLength*2)
	for len(code) < codeLength {
		if _, err := io.ReadFull(g.random, buffer); err != nil {
			return "", fmt.Errorf("generate short code: %w", err)
		}
		for _, value := range buffer {
			if int(value) >= maxByte {
				continue
			}
			code = append(code, alphabet[int(value)%len(alphabet)])
			if len(code) == codeLength {
				break
			}
		}
	}
	return string(code), nil
}
