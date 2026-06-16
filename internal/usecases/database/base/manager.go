package database

import (
	"context"

	errors "github.com/Scalingo/go-utils/errors/v3"
	scalingo "github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo"
	scalingobase "github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo/base"
	"github.com/Scalingo/scalingo-operator/internal/domain"
	"github.com/Scalingo/scalingo-operator/internal/usecases/database"
)

type manager struct {
	dbType   domain.DatabaseType
	scClient scalingo.Client
}

func NewManager(ctx context.Context, dbType domain.DatabaseType, apiToken, region string) (database.Manager, error) {
	err := dbType.Validate()
	if err != nil {
		return nil, errors.Wrap(ctx, err, "new manager")
	}
	if apiToken == "" {
		return nil, errors.New(ctx, "empty api token")
	}

	scClient, err := scalingobase.NewClient(ctx, apiToken, region)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "new scalingo client")
	}

	return &manager{
		dbType:   dbType,
		scClient: scClient,
	}, nil
}

func (m *manager) CreateDatabase(ctx context.Context, db domain.Database) (domain.Database, error) {
	return m.scClient.CreateDatabase(ctx, db)
}

// applyInstantDatabaseUpdates applies updates that do NOT require provisioning,
// such as firewall rules update.
// These updates are applied instantly or within few seconds.
func (m *manager) applyInstantDatabaseUpdates(ctx context.Context, db, expectedDB domain.Database) error {
	// Note: a `m.updateInternetAccess` implementation is available in this PR:
	// https://github.com/Scalingo/scalingo-operator/pull/22

	err := m.updateFirewallRules(ctx, db, expectedDB.FireWallRules)
	if err != nil {
		return errors.Wrap(ctx, err, "update firewall rules")
	}
	return nil
}

// updateDatabaseWithProvisioning applies one update at once that requires provisioning,
// such as plan change.
// These updates generally require minutes to be applied.
func (m *manager) updateDatabaseWithProvisioning(ctx context.Context, db, expectedDB domain.Database) (domain.DatabaseStatus, error) {
	if db.Status == domain.DatabaseStatusProvisioning {
		return db.Status, nil // Next updates can not occur while provisioning.
	}

	dbStatus, err := m.updateDatabasePlan(ctx, db, expectedDB)
	if err != nil {
		return db.Status, errors.Wrap(ctx, err, "update database plan")
	}
	return dbStatus, nil
}

func toDatabaseTypeName(ctx context.Context, dbType domain.DatabaseType) (string, error) {
	switch dbType {
	case domain.DatabaseTypePostgreSQL:
		return "POSTGRESQL", nil
	default:
		return "", errors.Newf(ctx, "no matching type for %q", dbType)
	}
}
