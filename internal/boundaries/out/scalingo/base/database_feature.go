package scalingo

import (
	"context"
	stderrors "errors"
	"fmt"

	scalingoapi "github.com/Scalingo/go-scalingo/v9"
	errors "github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func (c *client) EnableDatabaseFeature(ctx context.Context, db domain.Database, feature string) error {
	addonID, err := c.getAddonIDFromDatabase(ctx, db.Name)
	if err != nil {
		return errors.Wrap(ctx, err, "get database addon")
	}

	res, err := c.scClient.DatabaseEnableFeature(ctx, db.AppID, addonID, feature)
	if err != nil {
		return errors.Wrapf(ctx, err, "enable database feature %s", feature)
	}

	switch res.Status {
	case scalingoapi.DatabaseFeatureStatusActivated, scalingoapi.DatabaseFeatureStatusPending:
		return nil
	case scalingoapi.DatabaseFeatureStatusFailed:
		return errors.Wrapf(ctx, stderrors.New(res.Message), "enable database feature %s", feature)
	default:
		return fmt.Errorf("unknown database feature status %v", res.Status)
	}
}

func (c *client) DisableDatabaseFeature(ctx context.Context, db domain.Database, feature string) error {
	addonID, err := c.getAddonIDFromDatabase(ctx, db.Name)
	if err != nil {
		return errors.Wrap(ctx, err, "get database addon")
	}

	_, err = c.scClient.DatabaseDisableFeature(ctx, db.AppID, addonID, feature)
	if err != nil {
		return errors.Wrapf(ctx, err, "disable database feature %s", feature)
	}
	return nil
}

func toDatabaseFeatureStatus(status scalingoapi.DatabaseFeatureStatus) (domain.DatabaseFeatureStatus, error) {
	switch status {
	case scalingoapi.DatabaseFeatureStatusActivated:
		return domain.DatabaseFeatureStatusActivated, nil
	case scalingoapi.DatabaseFeatureStatusPending:
		return domain.DatabaseFeatureStatusPending, nil
	case scalingoapi.DatabaseFeatureStatusFailed:
		return domain.DatabaseFeatureStatusFailed, nil
	default:
		return domain.DatabaseFeatureStatus(""), fmt.Errorf("unknown database feature status %v", status)
	}
}

func toDatabaseFeatures(ctx context.Context, scFeatures []scalingoapi.DatabaseFeature) (domain.DatabaseFeatures, error) {
	features := make(domain.DatabaseFeatures, len(scFeatures))
	for _, scFeature := range scFeatures {
		status, err := toDatabaseFeatureStatus(scFeature.Status)
		if err != nil {
			return domain.DatabaseFeatures{}, errors.Wrap(ctx, err, "to database feature status")
		}
		features[scFeature.Name] = status
	}
	return features, nil
}
