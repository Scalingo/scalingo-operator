package helpers

import "fmt"

type DatabaseStatusCondition string

const (
	DatabaseStatusConditionAvailable    DatabaseStatusCondition = "Available"
	DatabaseStatusConditionProvisioning DatabaseStatusCondition = "Provisioning"

	// Databases status annotations.
	DatabaseAnnotationIsRunning = "databases.scalingo.com/db-is-running"
)

func (c DatabaseStatusCondition) Validate() error {
	switch c {
	case DatabaseStatusConditionAvailable, DatabaseStatusConditionProvisioning:
		return nil
	default:
		return fmt.Errorf("invalid database condition: %s", c)
	}
}
