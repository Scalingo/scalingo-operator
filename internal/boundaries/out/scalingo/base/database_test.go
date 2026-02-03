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
		status, err := toDatabaseStatus(scalingoapi.DatabaseStatusCreating)
		require.NoError(t, err)
		require.Equal(t, domain.DatabaseStatusProvisioning, status)

		status, err = toDatabaseStatus(scalingoapi.DatabaseStatusUpdating)
		require.NoError(t, err)
		require.Equal(t, domain.DatabaseStatusProvisioning, status)

		status, err = toDatabaseStatus(scalingoapi.DatabaseStatusMigrating)
		require.NoError(t, err)
		require.Equal(t, domain.DatabaseStatusProvisioning, status)

		status, err = toDatabaseStatus(scalingoapi.DatabaseStatusUpgrading)
		require.NoError(t, err)
		require.Equal(t, domain.DatabaseStatusProvisioning, status)
	})

	t.Run("it results as running status", func(t *testing.T) {
		status, err := toDatabaseStatus(scalingoapi.DatabaseStatusRunning)
		require.NoError(t, err)
		require.Equal(t, domain.DatabaseStatusRunning, status)
	})

	t.Run("it results as stopped status", func(t *testing.T) {
		status, err := toDatabaseStatus(scalingoapi.DatabaseStatusStopped)
		require.NoError(t, err)
		require.Equal(t, domain.DatabaseStatusStopped, status)
	})

	t.Run("it returns error for unknown API status", func(t *testing.T) {
		_, err := toDatabaseStatus("")
		require.ErrorContains(t, err, "unknown database status")

		_, err = toDatabaseStatus("whatever")
		require.ErrorContains(t, err, "unknown database status")
	})
}

func TestToDatabase(t *testing.T) {
	t.Run("it fails because of invalid db type", func(t *testing.T) {
		ctx := t.Context()
		db := scalingoapi.DatabaseNG{
			Database: scalingoapi.Database{
				ID:       "some_id",
				TypeName: "invalid_type",
			},
		}
		_, err := toDatabase(ctx, db)

		require.ErrorContains(t, err, "invalid database type")
	})

	t.Run("it fails because of unknown db status", func(t *testing.T) {
		ctx := t.Context()
		db := scalingoapi.DatabaseNG{
			Database: scalingoapi.Database{
				ID:       "some_id",
				TypeName: "postgresql",
				Status:   "unknown_status",
			},
		}
		_, err := toDatabase(ctx, db)

		require.ErrorContains(t, err, "unknown database status unknown_status")
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
				ID:       dbID,
				TypeName: "postgresql",
				Status:   scalingoapi.DatabaseStatusRunning,
			},
			App: scalingoapi.App{
				ID: appID,
			},
		}

		expectedDB := domain.Database{
			ID:       dbID,
			AppID:    appID,
			Name:     dbName,
			Type:     domain.DatabaseTypePostgreSQL,
			Status:   domain.DatabaseStatusRunning,
			Plan:     dbPlan,
			Features: domain.DatabaseFeatures{},
		}

		res, err := toDatabase(ctx, db)

		require.NoError(t, err)
		require.Equal(t, expectedDB, res)
	})
}
