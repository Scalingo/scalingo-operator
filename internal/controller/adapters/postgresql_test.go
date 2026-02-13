package adapters

import (
	"testing"

	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	databasesv1alpha1 "github.com/Scalingo/scalingo-operator/api/v1alpha1"

	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestPostgreSQLToDatabase(t *testing.T) {
	const (
		resourceName = "my-resource-name"
		dbName       = "my-dbng-name"
		dbPlan       = "postgresql-dr-enterprise-4096"
		projectID    = "prj-88888888-4444-4444-4444-cccccccccccc"
	)

	t.Run("it converts postgresql data from Kubebuilder to internal format", func(t *testing.T) {
		pg := databasesv1alpha1.PostgreSQL{
			Spec: databasesv1alpha1.PostgreSQLSpec{
				Name:      dbName,
				Plan:      dbPlan,
				ProjectID: projectID,
			},
		}
		expected := domain.Database{
			Name:      dbName,
			Type:      domain.DatabaseTypePostgreSQL,
			Plan:      dbPlan,
			ProjectID: projectID,
		}
		res, err := PostgreSQLToDatabase(t.Context(), pg)

		require.NoError(t, err)
		require.Equal(t, expected, res)
	})

	t.Run("it fallsback on resource name for missing database name", func(t *testing.T) {
		pg := databasesv1alpha1.PostgreSQL{
			ObjectMeta: metav1.ObjectMeta{
				Name: resourceName,
			},
			Spec: databasesv1alpha1.PostgreSQLSpec{
				Plan:      dbPlan,
				ProjectID: projectID,
			},
		}
		res, err := PostgreSQLToDatabase(t.Context(), pg)

		require.NoError(t, err)
		require.Equal(t, resourceName, res.Name)
	})
}
