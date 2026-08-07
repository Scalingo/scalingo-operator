package adapters

import (
	"context"

	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

// PostgreSQLToDatabase converts a PostgreSQL Kubernetes resource to the internal database representation.
func PostgreSQLToDatabase(ctx context.Context, postgresql apiv1.PostgreSQL) (domain.Database, error) {
	return toDatabase(
		ctx,
		domain.DatabaseTypePostgreSQL,
		postgresql.Name,
		postgresql.Spec.Name,
		postgresql.Spec.Plan,
		postgresql.Spec.ProjectID,
		postgresql.Spec.Networking,
	)
}
