package domain

import "fmt"

type DatabaseFeatureStatus string

const (
	DatabaseFeatureStatusActivated DatabaseFeatureStatus = "activated"
	DatabaseFeatureStatusPending   DatabaseFeatureStatus = "pending"
	DatabaseFeatureStatusFailed    DatabaseFeatureStatus = "failed"
)

func (s DatabaseFeatureStatus) Validate() error {
	switch s {
	case DatabaseFeatureStatusActivated, DatabaseFeatureStatusPending, DatabaseFeatureStatusFailed:
		return nil
	default:
		return fmt.Errorf("invalid database feature status: %s", s)
	}
}

// IsActive stands for activated, or going to be activated.
func (s DatabaseFeatureStatus) IsActive() bool {
	switch s {
	case DatabaseFeatureStatusActivated, DatabaseFeatureStatusPending:
		return true
	default:
		return false
	}
}
