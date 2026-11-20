package adapters

import (
	"testing"

	"github.com/stretchr/testify/require"

	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestValkeyToDatabase(t *testing.T) {
	const (
		resourceName = "my-resource-name"
		dbName       = "my-db-name"
		dbPlan       = "valkey-dr-tiny-starter-1024"
		projectID    = "prj-88888888-4444-4444-4444-cccccccccccc"
		ipRange      = "10.231.23.0/24"
	)

	t.Run("it converts valkey data from Kubebuilder to internal format", func(t *testing.T) {
		valkey := apiv1.Valkey{
			Spec: apiv1.ValkeySpec{
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
			Type:      domain.DatabaseTypeValkey,
			Plan:      dbPlan,
			ProjectID: projectID,
			IPRange:   ipRange,
		}

		res, err := ValkeyToDatabase(t.Context(), valkey)

		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	t.Run("it falls back on resource name for missing database name", func(t *testing.T) {
		valkey := apiv1.Valkey{
			Name: resourceName,
			Spec: apiv1.ValkeySpec{
				Plan:      dbPlan,
				ProjectID: projectID,
			},
		}

		res, err := ValkeyToDatabase(t.Context(), valkey)

		require.NoError(t, err)
		require.Equal(t, resourceName, res.Name)
	})
}
