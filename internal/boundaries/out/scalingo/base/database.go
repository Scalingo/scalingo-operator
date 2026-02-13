package scalingo

import (
	"context"

	scalingoapi "github.com/Scalingo/go-scalingo/v9"
	errors "github.com/Scalingo/go-utils/errors/v3"

	"github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo/base/adapters"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func (c *client) CreateDatabase(ctx context.Context, db domain.Database) (domain.Database, error) {

	addonProviderID, err := adapters.ToScalingoProviderId(db.Type)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "create database")
	}

	currentDB, err := c.scClient.Preview().DatabaseCreate(ctx, scalingoapi.DatabaseCreateParams{
		AddonProviderID: addonProviderID,
		PlanID:          db.Plan,
		Name:            db.Name,
		ProjectID:       db.ProjectID,
	})

	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "create database")
	}
	return adapters.ToDatabase(ctx, currentDB)
}

func (c *client) GetDatabase(ctx context.Context, dbID string) (domain.Database, error) {
	currentDB, err := c.scClient.Preview().DatabaseShow(ctx, dbID)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "get database")
	}

	addonID, err := c.getAddonIDFromDatabase(ctx, currentDB.Name)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "get database addon")
	}

	db, err := adapters.ToDatabase(ctx, currentDB)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "to database")
	}

	db.AddonID = addonID
	return db, nil
}

func (c *client) UpdateDatabasePlan(ctx context.Context, db domain.Database) error {
	currentDB, err := c.GetDatabase(ctx, db.ID)
	if err != nil {
		return errors.Wrap(ctx, err, "get database")
	}

	if db.Plan == currentDB.Plan {
		// Plan is expected to be checked before this method.
		return errors.Wrapf(ctx, domain.ErrNothingToBeDone, "already on plan %s", currentDB.Plan)
	}

	planID, err := c.findPlanID(ctx, currentDB.AddonID, db.Plan)
	if err != nil {
		return errors.Wrap(ctx, err, "invalid database plan")
	}

	_, err = c.scClient.AddonUpgrade(ctx, currentDB.ID, currentDB.AddonID, scalingoapi.AddonUpgradeParams{
		PlanID: planID,
	})
	if err != nil {
		return errors.Wrapf(ctx, err, "addon upgrade plan %s", db.Plan)
	}

	return nil
}

func (c *client) DeleteDatabase(ctx context.Context, dbID string) error {
	err := c.scClient.Preview().DatabaseDestroy(ctx, dbID)
	if err != nil {
		return errors.Wrap(ctx, err, "delete database")
	}
	return nil
}
