package database

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

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

func TestManager_CreateDatabase(t *testing.T) {
	t.Run("it successfully creates database", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		dbRequested := domain.Database{
			Name: "PG test",
			Type: domain.DatabaseTypePostgreSQL,
			Plan: "postgresql-dr-enterprise-4096",
		}

		dbCreated := dbRequested
		dbCreated.ID = databaseID

		scClient.EXPECT().CreateDatabase(ctx, dbRequested).Return(dbCreated, nil)

		// When
		res, err := manager.CreateDatabase(ctx, dbRequested)

		// Then
		require.NoError(t, err)
		require.Equal(t, dbCreated, res)
	})
}

func TestManager_GetDatabase(t *testing.T) {
	t.Run("it fails because of empty ID", func(t *testing.T) {
		ctx := t.Context()
		manager := manager{}
		res, err := manager.GetDatabase(ctx, "")

		require.EqualError(t, err, "empty database id")
		require.Empty(t, res)
	})

	t.Run("it successfully gets database", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		db := domain.Database{
			ID:      databaseID,
			AddonID: addonID,
			Name:    "PG test",
			Type:    domain.DatabaseTypePostgreSQL,
			Plan:    "postgresql-dr-enterprise-4096",
		}

		scClient.EXPECT().GetDatabase(ctx, databaseID).Return(db, nil)
		scClient.EXPECT().ListFirewallRules(ctx, databaseID, addonID)

		// When
		res, err := manager.GetDatabase(ctx, databaseID)

		// Then
		require.NoError(t, err)
		require.Equal(t, db, res)
	})
}

func TestManager_GetDatabaseURL(t *testing.T) {
	t.Run("it fails because of invalid datatype", func(t *testing.T) {
		ctx := t.Context()
		manager := manager{}
		db := domain.Database{
			Type: "invalid_type",
		}
		res, err := manager.GetDatabaseURL(ctx, db)

		require.EqualError(t, err, `to database type name: no matching type for "invalid_type"`)
		require.Empty(t, res)
	})

	t.Run("it successfully gets database URL", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		db := domain.Database{
			ID:    databaseID,
			AppID: appID,
			Type:  domain.DatabaseTypePostgreSQL,
		}

		dbURL := domain.DatabaseURL{
			Name:  "SCALINGO_POSTGRESQL_URL",
			Value: "postgresql_connection_string",
		}

		scClient.EXPECT().FindApplicationVariable(ctx, db.AppID, "SCALINGO_POSTGRESQL_URL").Return(dbURL.Value, nil)

		// When
		res, err := manager.GetDatabaseURL(ctx, db)

		// Then
		require.NoError(t, err)
		require.Equal(t, dbURL, res)
	})
}

func TestManager_DeleteDatabase(t *testing.T) {
	t.Run("it fails because of empty ID", func(t *testing.T) {
		ctx := t.Context()
		manager := manager{}
		err := manager.DeleteDatabase(ctx, "")

		require.EqualError(t, err, "empty database id")
	})

	t.Run("it successfully deletes database", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		scClient.EXPECT().DeleteDatabase(ctx, databaseID).Return(nil)

		// When
		err := manager.DeleteDatabase(ctx, databaseID)
		// Then
		require.NoError(t, err)
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
