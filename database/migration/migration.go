package migration

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var files embed.FS

// NewProvider returns a migration provider scoped to this database. The API
// deliberately does not call this package during startup.
func NewProvider(db *sql.DB) (*goose.Provider, error) {
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, files)
	if err != nil {
		return nil, fmt.Errorf("create goose provider: %w", err)
	}
	return provider, nil
}
