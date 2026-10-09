package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
)

var ErrInvalidHash = errors.New("invalid Argon2id hash")

type Parameters struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type Hasher struct {
	parameters Parameters
	random     io.Reader
}

func NewArgon2id() *Hasher {
	return &Hasher{
		parameters: Parameters{
			Memory:      19 * 1024,
			Iterations:  2,
			Parallelism: 1,
			SaltLength:  16,
			KeyLength:   32,
		},
		random: rand.Reader,
	}
}

func (h *Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.parameters.SaltLength)
	if _, err := io.ReadFull(h.random, salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.parameters.Iterations,
		h.parameters.Memory,
		h.parameters.Parallelism,
		h.parameters.KeyLength,
	)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.parameters.Memory,
		h.parameters.Iterations,
		h.parameters.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (h *Hasher) Verify(password, encodedHash string) (bool, error) {
	parameters, salt, expected, err := parseHash(encodedHash)
	if err != nil {
		return false, err
	}
	actual := argon2.IDKey(
		[]byte(password),
		salt,
		parameters.Iterations,
		parameters.Memory,
		parameters.Parallelism,
		uint32(len(expected)),
	)
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

func parseHash(encodedHash string) (Parameters, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Parameters{}, nil, nil, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return Parameters{}, nil, nil, ErrInvalidHash
	}

	var parameters Parameters
	if _, err := fmt.Sscanf(
		parts[3],
		"m=%d,t=%d,p=%d",
		&parameters.Memory,
		&parameters.Iterations,
		&parameters.Parallelism,
	); err != nil {
		return Parameters{}, nil, nil, ErrInvalidHash
	}
	if parameters.Memory < 8*1024 || parameters.Memory > 1024*1024 ||
		parameters.Iterations == 0 || parameters.Iterations > 10 ||
		parameters.Parallelism == 0 || parameters.Parallelism > 16 {
		return Parameters{}, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) < 8 || len(salt) > 64 {
		return Parameters{}, nil, nil, ErrInvalidHash
	}
	expected, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(expected) < 16 || len(expected) > 64 {
		return Parameters{}, nil, nil, ErrInvalidHash
	}
	return parameters, salt, expected, nil
}
