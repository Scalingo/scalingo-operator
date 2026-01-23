package scalingo

import (
	"context"
	"fmt"

	scalingoapi "github.com/Scalingo/go-scalingo/v9"
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

	addonID, err := c.getAddonIDFromDatabase(ctx, dbNG.Name)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "get database addon")
	}

	db, err := toDatabase(ctx, dbNG)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "to database")
	}

	db.AddonID = addonID
	return db, nil
}

// Shamelessly taken from `cli` project.
func (c *client) getAddonIDFromDatabase(ctx context.Context, databaseName string) (string, error) {
	// AddonsList works for both apps and DBNG databases (same API endpoint).
	// A DBNG database is modeled as an app with a single addon (itself),
	// whereas an application can have multiple addons (postgresql, redis, etc.).
	// If multiple addons are returned, the ID is likely an application, not a database.
	addons, err := c.scClient.AddonsList(ctx, databaseName)
	if err != nil {
		return "", errors.Wrap(ctx, err, "list addons")
	}

	if len(addons) == 0 {
		return "", errors.Newf(ctx, "no addon found for database %s", databaseName)
	}

	if len(addons) > 1 {
		return "", errors.Newf(ctx, "multiple addons found for %s, it may be an application", databaseName)
	}

	return addons[0].ID, nil
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

func toDatabaseStatus(status scalingoapi.DatabaseStatus) (domain.DatabaseStatus, error) {
	switch status {
	case scalingoapi.DatabaseStatusCreating, scalingoapi.DatabaseStatusUpdating,
		scalingoapi.DatabaseStatusMigrating, scalingoapi.DatabaseStatusUpgrading:
		return domain.DatabaseStatusProvisioning, nil
	case scalingoapi.DatabaseStatusRunning:
		return domain.DatabaseStatusRunning, nil
	case scalingoapi.DatabaseStatusStopped:
		return domain.DatabaseStatusStopped, nil
	default:
		return domain.DatabaseStatus(""), fmt.Errorf("unknown database status %v", status)
	}
}

func toDatabase(ctx context.Context, db scalingoapi.DatabaseNG) (domain.Database, error) {
	dbType := domain.DatabaseType(db.Database.TypeName)
	err := dbType.Validate()
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "to database type")
	}

	dbStatus, err := toDatabaseStatus(db.Database.Status)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "to database status")
	}

	return domain.Database{
		ID:        db.ID,
		AppID:     db.App.ID,
		Name:      db.Name,
		Type:      dbType,
		Status:    dbStatus,
		Plan:      db.Plan,
		ProjectID: db.ProjectID,
	}, nil
}
