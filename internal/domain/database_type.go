package domain

import "fmt"

type DatabaseType string

const (
	DatabaseTypeEmpty      DatabaseType = "" // Freshly created databases have no type defined
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
)

func (t DatabaseType) Validate() error {
	switch t {
	case DatabaseTypeEmpty, DatabaseTypePostgreSQL:
		return nil
	default:
		return fmt.Errorf("invalid database type: %s", t)
	}
}
