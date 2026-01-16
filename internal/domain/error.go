package domain

import "errors"

var (
	ErrNotImplemented   = errors.New("not implemented")
	ErrDatabaseNotFound = errors.New("database not found")
)
