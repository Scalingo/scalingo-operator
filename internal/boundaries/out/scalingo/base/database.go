package scalingo

import (
	"context"
	"fmt"

	scalingoapi "github.com/Scalingo/go-scalingo/v8"
	errors "github.com/Scalingo/go-utils/errors/v3"

	"github.com/Scalingo/scalingo-operator/internal/domain"
)

const postgresqlAddonProviderID = "postgresql-ng"

func (c *client) CreateDatabase(ctx context.Context, db domain.Database) (domain.Database, error) {
	addonProviderID, err := toScalingoProviderId(db.Type)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "create database")
	}

	dbNG, err := c.scClient.Preview().DatabaseCreate(ctx, scalingoapi.DatabaseCreateParams{
		AddonProviderID: addonProviderID,
		PlanID:          db.Plan,
		Name:            db.Name,
		ProjectID:       db.ProjectID,
	})

	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "create database")
	}
	return toDatabase(ctx, dbNG)
}

func (c *client) GetDatabase(ctx context.Context, dbID string) (domain.Database, error) {
	dbNG, err := c.scClient.Preview().DatabaseShow(ctx, dbID)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "get database")
	}

	return toDatabase(ctx, dbNG)
}

func (c *client) UpdateDatabase(ctx context.Context, db domain.Database) (domain.Database, error) {
	return domain.Database{}, domain.ErrNotImplemented
}

func (c *client) DeleteDatabase(ctx context.Context, dbID string) error {
	err := c.scClient.Preview().DatabaseDestroy(ctx, dbID)
	if err != nil {
		return errors.Wrap(ctx, err, "delete database")
	}
	return nil
}

func toScalingoProviderId(dbType domain.DatabaseType) (string, error) {
	switch dbType {
	case domain.DatabaseTypePostgreSQL:
		return postgresqlAddonProviderID, nil
	default:
		return "", fmt.Errorf("no matching provider for %q", dbType)
	}
}

func toDatabaseStatus(status scalingoapi.DatabaseStatus) domain.DatabaseStatus {
	switch status {

	case scalingoapi.DatabaseStatusCreating, scalingoapi.DatabaseStatusUpdating,
		scalingoapi.DatabaseStatusMigrating, scalingoapi.DatabaseStatusUpgrading:
		return domain.DatabaseStatusProvisioning
	case scalingoapi.DatabaseStatusRunning:
		return domain.DatabaseStatusRunning
	default:
		return domain.DatabaseStatusSuspended
	}
}
func toDatabase(ctx context.Context, db scalingoapi.DatabaseNG) (domain.Database, error) {
	dbType := domain.DatabaseType(db.Database.TypeName)
	err := dbType.Validate()
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "to database")
	}

	return domain.Database{
		ID:        db.ID,
		AppID:     db.App.ID,
		Name:      db.App.Name,
		Type:      dbType,
		Status:    toDatabaseStatus(db.Database.Status),
		Plan:      db.Database.Plan,
		ProjectID: db.App.Project.ID,
	}, nil
}
