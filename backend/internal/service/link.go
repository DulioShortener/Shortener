package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
)

const shortCodeAttempts = 5

type LinkRepository interface {
	Create(ctx context.Context, link entity.Link) (entity.Link, error)
	ListByUser(ctx context.Context, userID int64) ([]entity.Link, error)
	Delete(ctx context.Context, id, userID int64) error
	FindByCode(ctx context.Context, code string) (entity.Link, error)
}

type LinkService struct {
	links LinkRepository
	ids   IDGenerator
	codes CodeGenerator
	clock Clock
}

func NewLinkService(links LinkRepository, ids IDGenerator, codes CodeGenerator, clock Clock) *LinkService {
	return &LinkService{links: links, ids: ids, codes: codes, clock: clock}
}

func (s *LinkService) Create(ctx context.Context, userID int64, targetURL string) (entity.Link, error) {
	targetURL = strings.TrimSpace(targetURL)
	targetURLKey, err := normalizeURL(targetURL)
	if err != nil {
		return entity.Link{}, fmt.Errorf("normalize target URL: %w", err)
	}
	id, err := s.ids.NextID()
	if err != nil {
		return entity.Link{}, err
	}

	for range shortCodeAttempts {
		code, err := s.codes.Generate()
		if err != nil {
			return entity.Link{}, err
		}
		created, err := s.links.Create(ctx, entity.Link{
			ID: id, UserID: userID, Code: code, TargetURL: targetURL,
			TargetURLKey: targetURLKey, CreatedAt: s.clock.Now().UTC(),
		})
		if errors.Is(err, ErrShortCodeCollision) {
			continue
		}
		return created, err
	}
	return entity.Link{}, ErrShortCodeExhausted
}

func (s *LinkService) List(ctx context.Context, userID int64) ([]entity.Link, error) {
	return s.links.ListByUser(ctx, userID)
}

func (s *LinkService) Delete(ctx context.Context, id, userID int64) error {
	return s.links.Delete(ctx, id, userID)
}

func (s *LinkService) Resolve(ctx context.Context, code string) (entity.Link, error) {
	return s.links.FindByCode(ctx, code)
}

func normalizeURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	hostname := strings.ToLower(parsed.Hostname())
	port := parsed.Port()
	if (parsed.Scheme == "http" && port == "80") || (parsed.Scheme == "https" && port == "443") {
		port = ""
	}
	if strings.Contains(hostname, ":") {
		hostname = "[" + hostname + "]"
	}
	parsed.Host = hostname
	if port != "" {
		parsed.Host = net.JoinHostPort(strings.Trim(hostname, "[]"), port)
	}
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.String(), nil
}
