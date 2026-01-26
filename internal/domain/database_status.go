package domain

import "fmt"

type DatabaseStatus string

const (
	DatabaseStatusRunning      DatabaseStatus = "running"
	DatabaseStatusProvisioning DatabaseStatus = "provisioning"
	DatabaseStatusStopped      DatabaseStatus = "stopped"
)

func (s DatabaseStatus) Validate() error {
	switch s {
	case DatabaseStatusRunning, DatabaseStatusProvisioning, DatabaseStatusStopped:
		return nil
	default:
		return fmt.Errorf("invalid database status: %s", s)
	}
}
