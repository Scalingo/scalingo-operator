package adapters

import (
	"testing"

	"github.com/stretchr/testify/require"

	databasesv1alpha1 "github.com/Scalingo/scalingo-operator/api/v1alpha1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestPostgreSQLToDatabase(t *testing.T) {
	t.Run("it converts postgresql data from Kubebuilder to internal format", func(t *testing.T) {
		const (
			dbName    = "my-dbng-name"
			dbPlan    = "postgresql-ng-enterprise-4096"
			projectID = "prj-88888888-4444-4444-4444-cccccccccccc"
		)

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
}
