package domain

import "fmt"

type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeOpenSearch DatabaseType = "opensearch"
)

func (t DatabaseType) Validate() error {
	switch t {
	case DatabaseTypePostgreSQL, DatabaseTypeMySQL, DatabaseTypeOpenSearch:
		return nil
	default:
		return fmt.Errorf("invalid database type: %s", t)
	}
}
