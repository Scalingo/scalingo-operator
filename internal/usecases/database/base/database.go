package database

import (
	"context"

	"github.com/Scalingo/go-utils/errors/v2"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func (m *manager) updateDatabasePlan(ctx context.Context, db domain.Database, expectedDB domain.Database) (domain.DatabaseStatus, error) {
	if db.Plan == expectedDB.Plan {
		return db.Status, nil
	}

	if db.Status != domain.DatabaseStatusRunning {
		return db.Status, errors.Newf(ctx, "invalid status %s for plan update", db.Plan)
	}

	dbStatus, err := m.scClient.UpdateDatabasePlan(ctx, db, expectedDB.Plan)
	if err != nil {
		return db.Status, errors.Wrap(ctx, err, "update database plan")
	}

	return dbStatus, nil
}
