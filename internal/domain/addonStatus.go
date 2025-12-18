package domain

import "fmt"

type AddonStatus string

const (
	AddonStatusRunning      AddonStatus = "running"
	AddonStatusProvisioning AddonStatus = "provisioning"
	AddonStatusSuspended    AddonStatus = "suspended"
)

func (s AddonStatus) Validate() error {
	switch s {
	case AddonStatusRunning, AddonStatusProvisioning, AddonStatusSuspended:
		return nil
	default:
		return fmt.Errorf("invalid addon status: %s", s)
	}
}
