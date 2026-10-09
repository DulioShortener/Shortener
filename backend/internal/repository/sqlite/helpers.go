package sqlite

import (
	"errors"
	"strings"
	"time"

	gosqlite3 "github.com/mattn/go-sqlite3"
)

func unixMilliseconds(value time.Time) int64 {
	return value.UTC().UnixMilli()
}

func timeFromMilliseconds(value int64) time.Time {
	return time.UnixMilli(value).UTC()
}

func isConstraint(err error) bool {
	var sqliteError gosqlite3.Error
	return errors.As(err, &sqliteError) && sqliteError.Code == gosqlite3.ErrConstraint
}

func constraintContains(err error, value string) bool {
	return isConstraint(err) && strings.Contains(err.Error(), value)
}
