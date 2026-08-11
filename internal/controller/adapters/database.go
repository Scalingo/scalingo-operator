package adapters

import (
	"context"

	"github.com/Scalingo/go-utils/errors/v3"
	apiv1 "github.com/Scalingo/scalingo-operator/api/v1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func toDatabase(
	ctx context.Context,
	databaseType domain.DatabaseType,
	resourceName string,
	databaseName string,
	plan string,
	projectID string,
	networking apiv1.NetworkingSpec,
) (domain.Database, error) {
	rules, err := toFirewallRules(ctx, networking)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "to firewall rules")
	}

	if databaseName == "" {
		databaseName = resourceName
	}

	return domain.Database{
		Name:          databaseName,
		Type:          databaseType,
		Plan:          plan,
		ProjectID:     projectID,
		IPRange:       networking.IPRange,
		FireWallRules: rules,
	}, nil
}
