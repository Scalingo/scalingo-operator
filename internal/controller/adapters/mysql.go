package adapters

import (
	"context"

	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

// MySQLToDatabase converts a MySQL Kubernetes resource to the internal database representation.
func MySQLToDatabase(ctx context.Context, mysql apiv1.MySQL) (domain.Database, error) {
	return toDatabase(
		ctx,
		domain.DatabaseTypeMySQL,
		mysql.Name,
		mysql.Spec.Name,
		mysql.Spec.Plan,
		mysql.Spec.ProjectID,
		mysql.Spec.Networking,
	)
}
