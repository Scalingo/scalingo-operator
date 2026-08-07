package domain

import "fmt"

type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMySQL      DatabaseType = "mysql"
)

func (t DatabaseType) Validate() error {
	switch t {
	case DatabaseTypePostgreSQL, DatabaseTypeMySQL:
		return nil
	default:
		return fmt.Errorf("invalid database type: %s", t)
	}
}
