package adapters

import (
	"context"

	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

// ValkeyToDatabase converts a Valkey Kubernetes resource to the internal database representation.
func ValkeyToDatabase(ctx context.Context, valkey apiv1.Valkey) (domain.Database, error) {
	return toDatabase(
		ctx,
		domain.DatabaseTypeValkey,
		valkey.Name,
		valkey.Spec.Name,
		valkey.Spec.Plan,
		valkey.Spec.ProjectID,
		valkey.Spec.Networking,
	)
}
