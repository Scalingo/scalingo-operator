package adapters

import (
	"context"

	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

// OpenSearchToDatabase converts an OpenSearch Kubernetes resource to the internal database representation.
func OpenSearchToDatabase(ctx context.Context, opensearch apiv1.OpenSearch) (domain.Database, error) {
	return toDatabase(
		ctx,
		domain.DatabaseTypeOpenSearch,
		opensearch.Name,
		opensearch.Spec.Name,
		opensearch.Spec.Plan,
		opensearch.Spec.ProjectID,
		opensearch.Spec.Networking,
	)
}
