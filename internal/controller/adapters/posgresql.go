package adapters

import (
	databasesv1alpha1 "github.com/Scalingo/scalingo-operator/api/v1alpha1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

// Convert from Kubebuilder type to internal type.
func PostgreSQLToDatabase(postgresql databasesv1alpha1.PostgreSQL) domain.Database {
	return domain.Database{
		Name:      postgresql.Spec.Name,
		Type:      domain.DatabaseTypePostgreSQL,
		Plan:      postgresql.Spec.Plan,
		ProjectID: postgresql.Spec.ProjectID,
	}
}
