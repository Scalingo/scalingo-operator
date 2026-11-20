package domain

import "fmt"

type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeOpenSearch DatabaseType = "opensearch"
	DatabaseTypeValkey     DatabaseType = "valkey"
)

func (t DatabaseType) Validate() error {
	switch t {
	case DatabaseTypePostgreSQL, DatabaseTypeMySQL, DatabaseTypeOpenSearch, DatabaseTypeValkey:
		return nil
	default:
		return fmt.Errorf("invalid database type: %s", t)
	}
}
