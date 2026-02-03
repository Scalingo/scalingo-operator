package adapters

import (
	"context"

	"github.com/Scalingo/go-utils/errors/v3"

	apiv1alpha1 "github.com/Scalingo/scalingo-operator/api/v1alpha1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

// Convert from Kubebuilder type to internal type.
func PostgreSQLToDatabase(ctx context.Context, postgresql apiv1alpha1.PostgreSQL) (domain.Database, error) {
	rules, err := toFirewallRules(ctx, postgresql.Spec.Networking)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "to firewall rules")
	}

	return domain.Database{
		Name:          postgresql.Spec.Name,
		Type:          domain.DatabaseTypePostgreSQL,
		Plan:          postgresql.Spec.Plan,
		ProjectID:     postgresql.Spec.ProjectID,
		Features:      toFeatures(postgresql.Spec.Networking),
		FireWallRules: rules,
	}, nil
}
