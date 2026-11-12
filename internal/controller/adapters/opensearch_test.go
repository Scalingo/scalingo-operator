package adapters

import (
	"testing"

	"github.com/stretchr/testify/require"

	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestOpenSearchToDatabase(t *testing.T) {
	const (
		resourceName = "my-resource-name"
		dbName       = "my-db-name"
		dbPlan       = "opensearch-dr-starter-8192"
		projectID    = "prj-88888888-4444-4444-4444-cccccccccccc"
		ipRange      = "10.231.23.0/24"
	)

	t.Run("it converts opensearch data from Kubebuilder to internal format", func(t *testing.T) {
		opensearch := apiv1.OpenSearch{
			Spec: apiv1.OpenSearchSpec{
				Name: dbName,
				Networking: apiv1.NetworkingSpec{
					IPRange: ipRange,
				},
				Plan:      dbPlan,
				ProjectID: projectID,
			},
		}
		expected := domain.Database{
			Name:      dbName,
			Type:      domain.DatabaseTypeOpenSearch,
			Plan:      dbPlan,
			ProjectID: projectID,
			IPRange:   ipRange,
		}

		res, err := OpenSearchToDatabase(t.Context(), opensearch)

		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	t.Run("it falls back on resource name for missing database name", func(t *testing.T) {
		opensearch := apiv1.OpenSearch{
			Name: resourceName,
			Spec: apiv1.OpenSearchSpec{
				Plan:      dbPlan,
				ProjectID: projectID,
			},
		}

		res, err := OpenSearchToDatabase(t.Context(), opensearch)

		require.NoError(t, err)
		require.Equal(t, resourceName, res.Name)
	})
}
