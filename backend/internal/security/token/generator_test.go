package token

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

func TestGenerateReturnsRawTokenAndVerifier(t *testing.T) {
	generator := &Generator{random: bytes.NewReader(make([]byte, tokenBytes))}
	generated, err := generator.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if generated.Raw == "" || len(generated.Hash) != sha256.Size {
		t.Fatalf("unexpected token: raw=%q hash length=%d", generated.Raw, len(generated.Hash))
	}
	if !bytes.Equal(generated.Hash, generator.Hash(generated.Raw)) {
		t.Fatal("stored verifier does not match raw token")
	}
}
