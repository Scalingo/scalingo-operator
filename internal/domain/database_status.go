package domain

import "fmt"

type DatabaseStatus string

const (
	DatabaseStatusRunning      DatabaseStatus = "running"
	DatabaseStatusProvisioning DatabaseStatus = "provisioning"
	DatabaseStatusSuspended    DatabaseStatus = "suspended"
)

func (s DatabaseStatus) Validate() error {
	switch s {
	case DatabaseStatusRunning, DatabaseStatusProvisioning, DatabaseStatusSuspended:
		return nil
	default:
		return fmt.Errorf("invalid database status: %s", s)
	}
}
