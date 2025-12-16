package domain

import "errors"

var (
	ErrNotImplemented = errors.New("not implemented") // TODO(david): remove before first Prod release

	ErrDatabaseNotFound = errors.New("database not found")
)
