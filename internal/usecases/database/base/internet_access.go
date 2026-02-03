package database

import (
	"context"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

// updateInternetAccess enables or disables public internet access using the database features.
//
// The feature to enable the database internet access is "publicly-available".
// It requires "force-ssl" feature.
// More info:
// https://developers.scalingo.com/databases/databases#enable-feature
func (m *manager) updateInternetAccess(ctx context.Context, currentDB domain.Database, expectedEnable bool) error {
	currentForceSsl, currentEnable := extractInternetAccessFeatures(currentDB)
	if currentEnable == expectedEnable {
		return nil
	}

	log := logf.FromContext(ctx)

	if expectedEnable && !currentForceSsl {
		log.Info("Enable database feature", "feature", domain.DatabaseFeatureForceSsl)

		err := m.scClient.EnableDatabaseFeature(ctx, currentDB, domain.DatabaseFeatureForceSsl)
		if err != nil {
			return errors.Wrapf(ctx, err, "enable feature %s", domain.DatabaseFeatureForceSsl)
		}
	}

	if expectedEnable {
		log.Info("Enable database feature", "feature", domain.DatabaseFeaturePubliclyAvailable)

		err := m.scClient.EnableDatabaseFeature(ctx, currentDB, domain.DatabaseFeaturePubliclyAvailable)
		if err != nil {
			return errors.Wrapf(ctx, err, "enable feature %s", domain.DatabaseFeaturePubliclyAvailable)
		}
	} else {
		log.Info("Disable database feature", "feature", domain.DatabaseFeaturePubliclyAvailable)

		err := m.scClient.DisableDatabaseFeature(ctx, currentDB, domain.DatabaseFeaturePubliclyAvailable)
		if err != nil {
			return errors.Wrapf(ctx, err, "disable feature %s", domain.DatabaseFeaturePubliclyAvailable)
		}
	}
	return nil
}

// extractIntenertAccessFeatures returns extracted forceSsl and enabled features.
func extractInternetAccessFeatures(currentDB domain.Database) (bool, bool) {
	if len(currentDB.Features) == 0 {
		return false, false
	}

	status, ok := currentDB.Features[domain.DatabaseFeatureForceSsl]
	forceSsl := ok && status.IsActive()

	status, ok = currentDB.Features[domain.DatabaseFeaturePubliclyAvailable]
	enable := ok && status.IsActive()

	return forceSsl, enable
}
