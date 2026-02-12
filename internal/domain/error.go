package domain

import "errors"

var (
	ErrNotImplemented   = errors.New("not implemented")
	ErrDatabaseNotFound = errors.New("database not found")
	ErrNothingToBeDone  = errors.New("nothing to be done")

	// Not properly an error, means the result is in progress, not done yet.
	ErrProvisioning = errors.New("provisioning")
)
