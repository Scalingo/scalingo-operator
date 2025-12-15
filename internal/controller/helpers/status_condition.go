package helpers

import "fmt"

type DatabaseStatusCondition string

const (
	DatabaseStatusConditionAvailable    DatabaseStatusCondition = "Available"
	DatabaseStatusConditionProvisioning DatabaseStatusCondition = "Provisioning"
)

func (c DatabaseStatusCondition) Validate() error {
	switch c {
	case DatabaseStatusConditionAvailable, DatabaseStatusConditionProvisioning:
		return nil
	default:
		return fmt.Errorf("invalid database status condition: %s", c)
	}
}
