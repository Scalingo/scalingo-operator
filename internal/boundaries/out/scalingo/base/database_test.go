package scalingo

import (
	"testing"

	"github.com/stretchr/testify/require"

	scalingoapi "github.com/Scalingo/go-scalingo/v9"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestToScalingoProviderId(t *testing.T) {
	tests := map[string]struct {
		dbType             domain.DatabaseType
		scalingoProviderId string
		isExpectedError    bool
	}{
		"it fails with empty type": {
			dbType:          domain.DatabaseType(""),
			isExpectedError: true,
		},
		"it fails with unknown type": {
			dbType:          domain.DatabaseType("unknown"),
			isExpectedError: true,
		},
		"it successfully extract PostgreSQL Scalingo addon provier ID": {
			dbType:             domain.DatabaseTypePostgreSQL,
			scalingoProviderId: postgresqlAddonProviderID,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := toScalingoProviderId(test.dbType)

			if test.isExpectedError {
				require.ErrorContains(t, err, "no matching provider for ")
			} else {
				require.NoError(t, err)
				require.Equal(t, test.scalingoProviderId, res)
			}
		})
	}
}

func TestToScalingoStatus(t *testing.T) {
	t.Run("it results as provisioning status", func(t *testing.T) {
		require.Equal(t, domain.DatabaseStatusProvisioning, toDatabaseStatus(scalingoapi.DatabaseStatusCreating))
		require.Equal(t, domain.DatabaseStatusProvisioning, toDatabaseStatus(scalingoapi.DatabaseStatusUpdating))
		require.Equal(t, domain.DatabaseStatusProvisioning, toDatabaseStatus(scalingoapi.DatabaseStatusMigrating))
		require.Equal(t, domain.DatabaseStatusProvisioning, toDatabaseStatus(scalingoapi.DatabaseStatusUpgrading))
	})

	t.Run("it results as running status", func(t *testing.T) {
		require.Equal(t, domain.DatabaseStatusRunning, toDatabaseStatus(scalingoapi.DatabaseStatusRunning))
	})

	t.Run("it falls back on suspended status for unknow API status", func(t *testing.T) {
		require.Equal(t, domain.DatabaseStatusSuspended, toDatabaseStatus(""))
		require.Equal(t, domain.DatabaseStatusSuspended, toDatabaseStatus("unknown"))
	})
}

func TestToDatabase(t *testing.T) {
	t.Run("it fails because of invalid db type", func(t *testing.T) {
		ctx := t.Context()
		db := scalingoapi.DatabaseNG{
			Database: scalingoapi.Database{
				TypeName: "invalid_type",
			},
		}
		_, err := toDatabase(ctx, db)

		require.ErrorContains(t, err, "invalid database type")
	})

	t.Run("it converts to database", func(t *testing.T) {
		ctx := t.Context()

		const (
			dbID   = "db_id"
			dbName = "db_name"
			dbPlan = "db_plan_name"
			appID  = "app_id"
		)

		db := scalingoapi.DatabaseNG{
			ID:   dbID,
			Name: dbName,
			Plan: dbPlan,
			Database: scalingoapi.Database{
				TypeName: "postgresql",
				Status:   scalingoapi.DatabaseStatusRunning,
			},
			App: scalingoapi.App{
				ID: appID,
			},
		}

		expectedDB := domain.Database{
			ID:     dbID,
			AppID:  appID,
			Name:   dbName,
			Type:   domain.DatabaseTypePostgreSQL,
			Status: domain.DatabaseStatusRunning,
			Plan:   dbPlan,
		}

		res, err := toDatabase(ctx, db)

		require.NoError(t, err)
		require.Equal(t, expectedDB, res)
	})
}
