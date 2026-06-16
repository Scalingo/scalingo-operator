package database

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo/scalingomock"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

const (
	databaseID = "db_test_id"
	appID      = "app_test_id"
	addonID    = "addon_test_id"
)

func TestNewManager(t *testing.T) {
	t.Run("it fails because of bad database type", func(t *testing.T) {
		ctx := t.Context()
		dbManager, err := NewManager(ctx, "invalid_db_type", "", "")

		require.EqualError(t, err, "new manager: invalid database type: invalid_db_type")
		require.Nil(t, dbManager)
	})

	t.Run("it fails because of empty API token", func(t *testing.T) {
		ctx := t.Context()
		dbManager, err := NewManager(ctx, domain.DatabaseTypePostgreSQL, "", "")

		require.EqualError(t, err, "empty api token")
		require.Nil(t, dbManager)
	})
}

func TestManager_updateDatabaseWithProvisioning(t *testing.T) {
	t.Run("it does nothing if database is provisioning", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			Status: domain.DatabaseStatusProvisioning,
			Plan:   "postgresql-dr-enterprise-4096",
		}

		expectedDB := domain.Database{
			Plan: "postgresql-dr-enterprise-8192",
		}

		// When
		status, err := manager.updateDatabaseWithProvisioning(ctx, currentDB, expectedDB)

		// Then
		require.NoError(t, err)
		require.Equal(t, domain.DatabaseStatusProvisioning, status)
	})

	t.Run("it updates the database if not provisioning", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			Status: domain.DatabaseStatusRunning,
			Plan:   "postgresql-dr-enterprise-4096",
		}

		expectedDB := domain.Database{
			Plan: "postgresql-dr-enterprise-8192",
		}

		scClient.EXPECT().UpdateDatabasePlan(ctx, currentDB, expectedDB.Plan).
			Return(domain.DatabaseStatusProvisioning, nil)

		// When
		status, err := manager.updateDatabaseWithProvisioning(ctx, currentDB, expectedDB)

		// Then
		require.NoError(t, err)
		require.Equal(t, domain.DatabaseStatusProvisioning, status)
	})
}
func TestToDatabaseTypeName(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		dbType          domain.DatabaseType
		dbTypeName      string
		isExpectedError bool
	}{
		"it fails with empty type": {
			dbType:          domain.DatabaseType(""),
			isExpectedError: true,
		},
		"it fails with unknown type": {
			dbType:          domain.DatabaseType("unknown"),
			isExpectedError: true,
		},
		"it successfully extract PostgreSQL type name": {
			dbType:     domain.DatabaseTypePostgreSQL,
			dbTypeName: "POSTGRESQL",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := toDatabaseTypeName(ctx, test.dbType)

			if test.isExpectedError {
				require.ErrorContains(t, err, "no matching type for ")
			} else {
				require.NoError(t, err)
				require.Equal(t, test.dbTypeName, res)
			}
		})
	}

	t.Run("it fails because of bad database type", func(t *testing.T) {
		ctx := t.Context()
		dbManager, err := NewManager(ctx, "invalid_db_type", "", "")

		require.EqualError(t, err, "new manager: invalid database type: invalid_db_type")
		require.Nil(t, dbManager)
	})
}
