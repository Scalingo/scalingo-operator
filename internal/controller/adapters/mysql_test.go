package adapters

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestMySQLToDatabase(t *testing.T) {
	const (
		resourceName = "my-resource-name"
		dbName       = "my-db-name"
		dbPlan       = "mysql-dr-starter-8192"
		projectID    = "prj-88888888-4444-4444-4444-cccccccccccc"
		ipRange      = "10.231.23.0/24"
	)

	t.Run("it converts mysql data from Kubebuilder to internal format", func(t *testing.T) {
		mysql := apiv1.MySQL{
			Spec: apiv1.MySQLSpec{
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
			Type:      domain.DatabaseTypeMySQL,
			Plan:      dbPlan,
			ProjectID: projectID,
			IPRange:   ipRange,
		}

		res, err := MySQLToDatabase(t.Context(), mysql)

		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	t.Run("it falls back on resource name for missing database name", func(t *testing.T) {
		mysql := apiv1.MySQL{
			ObjectMeta: metav1.ObjectMeta{Name: resourceName},
			Spec: apiv1.MySQLSpec{
				Plan:      dbPlan,
				ProjectID: projectID,
			},
		}

		res, err := MySQLToDatabase(t.Context(), mysql)

		require.NoError(t, err)
		require.Equal(t, resourceName, res.Name)
	})
}
