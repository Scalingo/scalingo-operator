package database

import (
	"context"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	errors "github.com/Scalingo/go-utils/errors/v3"
	scalingobase "github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo/base"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func (m *manager) CheckDatabaseExists(ctx context.Context, dbID string) (bool, error) {
	_, err := m.scClient.GetDatabase(ctx, dbID)
	if errors.Is(err, scalingobase.ErrDatabaseNotFound) {
		return false, nil
	} else if err != nil {
		return false, errors.Wrapf(ctx, err, "check database %s exists", dbID)
	}
	return true, nil
}

func (m *manager) GetDatabase(ctx context.Context, dbID string) (domain.Database, error) {
	log := logf.FromContext(ctx)

	if dbID == "" {
		return domain.Database{}, errors.New(ctx, "empty database id")
	}
	db, err := m.scClient.GetDatabase(ctx, dbID)
	if err != nil {
		return domain.Database{}, errors.Wrapf(ctx, err, "get database %s", dbID)
	}

	rules, err := m.scClient.ListFirewallRules(ctx, db.ID, db.AddonID)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "get database firewall rules")
	}

	db.FireWallRules = rules
	log.Info("Get database", "database", db)

	return db, nil
}

func (m *manager) GetDatabaseURL(ctx context.Context, db domain.Database) (domain.DatabaseURL, error) {
	dbTypeName, err := toDatabaseTypeName(ctx, db.Type)
	if err != nil {
		return domain.DatabaseURL{}, errors.Wrap(ctx, err, "to database type name")
	}

	varName := "SCALINGO_" + dbTypeName + "_URL"
	varValue, err := m.scClient.FindApplicationVariable(ctx, db.AppID, varName)
	if err != nil {
		return domain.DatabaseURL{}, errors.Wrap(ctx, err, "find database url")
	}
	return domain.DatabaseURL{
		Name:  varName,
		Value: varValue,
	}, nil
}

func (m *manager) UpdateDatabase(ctx context.Context, dbID string, expectedDB domain.Database) (domain.DatabaseStatus, error) {
	db, err := m.GetDatabase(ctx, dbID)
	if err != nil {
		return domain.DatabaseStatusUnknown, errors.Wrapf(ctx, err, "get database %s", dbID)
	}

	err = m.applyInstantDatabaseUpdates(ctx, db, expectedDB)
	if err != nil {
		return db.Status, err
	}

	return m.updateDatabaseWithProvisioning(ctx, db, expectedDB)
}

func (m *manager) updateDatabasePlan(ctx context.Context, db domain.Database, expectedDB domain.Database) (domain.DatabaseStatus, error) {
	if db.Plan == expectedDB.Plan {
		return db.Status, nil
	}

	if db.Status != domain.DatabaseStatusRunning {
		return db.Status, errors.Newf(ctx, "invalid status %s for plan update", db.Status)
	}

	dbStatus, err := m.scClient.UpdateDatabasePlan(ctx, db, expectedDB.Plan)
	if err != nil {
		return db.Status, errors.Wrap(ctx, err, "update database plan")
	}

	return dbStatus, nil
}

func (m *manager) DeleteDatabase(ctx context.Context, dbID string) error {
	if dbID == "" {
		return errors.New(ctx, "empty database id")
	}

	err := m.scClient.DeleteDatabase(ctx, dbID)
	if err != nil {
		return errors.Wrapf(ctx, err, "delete database %v", dbID)
	}
	return nil
}
