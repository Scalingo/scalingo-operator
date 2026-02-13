package adapters

import (
	"context"
	"fmt"

	scalingoapi "github.com/Scalingo/go-scalingo/v9"
	errors "github.com/Scalingo/go-utils/errors/v3"

	"github.com/Scalingo/scalingo-operator/internal/domain"
)

const postgresqlAddonProviderID = "postgresql-ng"

func ToScalingoProviderId(dbType domain.DatabaseType) (string, error) {
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

func ToDatabase(ctx context.Context, db scalingoapi.DatabaseNG) (domain.Database, error) {
	var dbType domain.DatabaseType
	var dbStatus domain.DatabaseStatus

	// Freshly created databases come with empty Database subobject.
	// There is neither type nor status to read from empty Database.
	if db.Database.ID != "" {
		dbType = domain.DatabaseType(db.Database.TypeName)
		err := dbType.Validate()
		if err != nil {
			return domain.Database{}, errors.Wrap(ctx, err, "to database type")
		}

		dbStatus, err = toDatabaseStatus(db.Database.Status)
		if err != nil {
			return domain.Database{}, errors.Wrap(ctx, err, "to database status")
		}
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
