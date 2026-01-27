package domain

import "fmt"

type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
)

func (t DatabaseType) Validate() error {
	switch t {
	case DatabaseTypePostgreSQL:
		return nil
	default:
		return fmt.Errorf("invalid database type: %s", t)
	}
}
