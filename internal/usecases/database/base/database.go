package database

import (
	"context"

	"github.com/Scalingo/go-utils/errors/v2"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func (m *manager) updateDatabasePlan(ctx context.Context, currentDB domain.Database, expectedDB domain.Database) error {
	if currentDB.Plan == expectedDB.Plan {
		return nil
	}

	if currentDB.Status != domain.DatabaseStatusRunning {
		return errors.Newf(ctx, "invalid status %s for plan update", currentDB.Plan)
	}

	err := m.scClient.UpdateDatabasePlan(ctx, expectedDB)
	if err != nil {
		return errors.Wrap(ctx, err, "update database plan")
	}

	return domain.ErrProvisioning
}
