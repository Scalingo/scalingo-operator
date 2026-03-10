package database

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo/scalingomock"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestManager_updateDatabasePlan(t *testing.T) {
	t.Run("it does nothing if current plan is the same as expected", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID:     databaseID,
			Plan:   "postgresql-dr-enterprise-4096",
			Status: domain.DatabaseStatusRunning,
		}
		expectedDB := domain.Database{Plan: "postgresql-dr-enterprise-4096"}

		dbStatus, err := manager.updateDatabasePlan(ctx, currentDB, expectedDB)
		require.NoError(t, err)
		require.Equal(t, currentDB.Status, dbStatus)
	})

	t.Run("it returns error if database status is not running", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID:     databaseID,
			Plan:   "postgresql-dr-enterprise-2048",
			Status: domain.DatabaseStatusProvisioning,
		}
		expectedDB := domain.Database{Plan: "postgresql-dr-enterprise-4096"}

		_, err := manager.updateDatabasePlan(ctx, currentDB, expectedDB)
		require.ErrorContains(t, err, "invalid status")
	})

	t.Run("it successfully updates database plan", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID:     databaseID,
			Plan:   "postgresql-dr-enterprise-2048",
			Status: domain.DatabaseStatusRunning,
		}
		expectedDB := domain.Database{Plan: "postgresql-dr-enterprise-4096"}

		scClient.EXPECT().UpdateDatabasePlan(ctx, currentDB, expectedDB.Plan).Return(domain.DatabaseStatusProvisioning, nil)

		dbStatus, err := manager.updateDatabasePlan(ctx, currentDB, expectedDB)
		require.NoError(t, err)
		require.Equal(t, domain.DatabaseStatusProvisioning, dbStatus)
	})

	t.Run("it returns error if scClient.UpdateDatabasePlan fails", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			Plan:   "postgresql-dr-enterprise-2048",
			Status: domain.DatabaseStatusRunning,
		}
		expectedDB := domain.Database{Plan: "postgresql-dr-enterprise-4096"}

		scClient.EXPECT().UpdateDatabasePlan(ctx, currentDB, expectedDB.Plan).Return(currentDB.Status, errors.New("provisioning"))

		_, err := manager.updateDatabasePlan(ctx, currentDB, expectedDB)
		require.Error(t, err)
		require.ErrorContains(t, err, "update database plan")
	})
}
