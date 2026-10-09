package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DulioShortener/Shortener/backend/internal/entity"
)

type linkRepositoryStub struct {
	createCalls int
	created     entity.Link
}

func (r *linkRepositoryStub) Create(_ context.Context, link entity.Link) (entity.Link, error) {
	r.createCalls++
	if r.createCalls == 1 {
		return entity.Link{}, ErrShortCodeCollision
	}
	r.created = link
	return link, nil
}

func (*linkRepositoryStub) ListByUser(context.Context, int64) ([]entity.Link, error) {
	return nil, nil
}

func (*linkRepositoryStub) Delete(context.Context, int64, int64) error { return nil }

func (*linkRepositoryStub) FindByCode(context.Context, string) (entity.Link, error) {
	return entity.Link{}, errors.New("not used")
}

type codeSequence struct {
	codes []string
	index int
}

func (g *codeSequence) Generate() (string, error) {
	value := g.codes[g.index]
	g.index++
	return value, nil
}

func TestLinkCreateNormalizesDuplicateKeyAndRetriesCodeCollision(t *testing.T) {
	repository := &linkRepositoryStub{}
	clock := fixedClock(time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC))
	links := NewLinkService(
		repository, fixedID(42), &codeSequence{codes: []string{"AAAAAAAA", "BBBBBBBB"}}, clock,
	)
	created, err := links.Create(context.Background(), 7, "https://Example.com:443")
	if err != nil {
		t.Fatal(err)
	}
	if repository.createCalls != 2 || created.Code != "BBBBBBBB" {
		t.Fatalf("expected collision retry, calls=%d code=%q", repository.createCalls, created.Code)
	}
	if created.TargetURLKey != "https://example.com/" {
		t.Fatalf("unexpected canonical URL key: %q", created.TargetURLKey)
	}
}
