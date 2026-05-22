package database

import (
	"errors"
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

func TestManager_GetDatabaseNetworkConfiguration(t *testing.T) {
	t.Run("it fails because of empty ID", func(t *testing.T) {
		ctx := t.Context()
		manager := manager{}
		res, err := manager.GetDatabaseNetworkConfiguration(ctx, "")

		require.EqualError(t, err, "empty database id")
		require.Empty(t, res)
	})

	t.Run("it successfully gets database network configuration", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		config := domain.DatabaseNetworkConfiguration{
			OutscaleAccountID: "123456789012",
			OutscaleNetID:     "vpc-1234abcd",
			IPRange:           "10.0.0.0/24",
		}

		scClient.EXPECT().GetDatabaseNetworkConfiguration(ctx, databaseID).Return(config, nil)

		res, err := manager.GetDatabaseNetworkConfiguration(ctx, databaseID)

		require.NoError(t, err)
		require.Equal(t, config, res)
	})
}

func TestManager_GetDatabaseEndpoints(t *testing.T) {
	t.Run("fails because of empty ID", func(t *testing.T) {
		ctx := t.Context()
		manager := manager{}
		res, err := manager.GetDatabaseEndpoints(ctx, "")

		require.EqualError(t, err, "empty database id")
		require.Empty(t, res)
	})

	t.Run("successfully gets database endpoints", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		endpoints := []domain.DatabaseEndpoint{
			{
				ID:       "endpoint-1",
				Hostname: "public-host",
				Port:     5432,
				Type:     domain.DatabaseEndpointTypePublicRW,
			},
		}

		scClient.EXPECT().ListDatabaseEndpoints(ctx, databaseID).Return(endpoints, nil)

		res, err := manager.GetDatabaseEndpoints(ctx, databaseID)

		require.NoError(t, err)
		require.Equal(t, endpoints, res)
	})
}

func TestManager_EnsureDatabaseNetPeering(t *testing.T) {
	t.Run("it fails because of empty database ID", func(t *testing.T) {
		ctx := t.Context()
		manager := manager{}
		err := manager.EnsureDatabaseNetPeering(ctx, "", "pcx-1234")

		require.EqualError(t, err, "empty database id")
	})

	t.Run("it fails because of empty net peering ID", func(t *testing.T) {
		ctx := t.Context()
		manager := manager{}
		err := manager.EnsureDatabaseNetPeering(ctx, databaseID, "")

		require.EqualError(t, err, "empty outscale net peering id")
	})

	t.Run("it does nothing when matching net peering already exists", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		scClient.EXPECT().ListDatabaseNetPeerings(ctx, databaseID).Return([]domain.DatabaseNetPeering{
			{
				ID:                   "np-1",
				OutscaleNetPeeringID: "pcx-1234",
			},
		}, nil)

		err := manager.EnsureDatabaseNetPeering(ctx, databaseID, "pcx-1234")
		require.NoError(t, err)
	})

	t.Run("it creates database net peering when missing", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		scClient.EXPECT().ListDatabaseNetPeerings(ctx, databaseID).Return([]domain.DatabaseNetPeering{}, nil)
		scClient.EXPECT().CreateDatabaseNetPeering(ctx, databaseID, "pcx-1234").Return(domain.DatabaseNetPeering{
			ID:                   "np-1",
			OutscaleNetPeeringID: "pcx-1234",
		}, nil)

		err := manager.EnsureDatabaseNetPeering(ctx, databaseID, "pcx-1234")
		require.NoError(t, err)
	})
}

func TestManager_DeleteDatabaseNetPeering(t *testing.T) {
	t.Run("fails because of empty database ID", func(t *testing.T) {
		ctx := t.Context()
		manager := manager{}
		err := manager.DeleteDatabaseNetPeering(ctx, "", "np-1")

		require.EqualError(t, err, "empty database id")
	})

	t.Run("fails because of empty net peering ID", func(t *testing.T) {
		ctx := t.Context()
		manager := manager{}
		err := manager.DeleteDatabaseNetPeering(ctx, databaseID, "")

		require.EqualError(t, err, "empty net peering id")
	})

	t.Run("deletes database net peering", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		scClient.EXPECT().DeleteDatabaseNetPeering(ctx, databaseID, "np-1").Return(nil)

		err := manager.DeleteDatabaseNetPeering(ctx, databaseID, "np-1")
		require.NoError(t, err)
	})
}

func TestManager_UpdateDatabase(t *testing.T) {
	t.Run("it fails when getting database", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		scClient.EXPECT().GetDatabase(ctx, databaseID).Return(domain.Database{}, errors.New("boom"))

		// When
		status, err := manager.UpdateDatabase(ctx, databaseID, domain.Database{})

		// Then
		require.Equal(t, domain.DatabaseStatusUnknown, status)
		require.ErrorContains(t, err, "get database")
	})

	t.Run("it updates firewall rules only when database is provisioning", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID:      databaseID,
			AddonID: addonID,
			Plan:    "postgresql-dr-starter-4096",
			Status:  domain.DatabaseStatusProvisioning,
		}
		expectedDB := domain.Database{
			Plan: "postgresql-dr-enterprise-4096",
		}

		scClient.EXPECT().GetDatabase(ctx, databaseID).Return(currentDB, nil)
		scClient.EXPECT().ListFirewallRules(ctx, databaseID, addonID).Return(nil, nil)

		// When
		status, err := manager.UpdateDatabase(ctx, databaseID, expectedDB)

		// Then
		require.NoError(t, err)
		require.Equal(t, domain.DatabaseStatusProvisioning, status)
	})

	t.Run("it updates plan when database is running", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID:      databaseID,
			AddonID: addonID,
			Plan:    "postgresql-dr-enterprise-2048",
			Status:  domain.DatabaseStatusRunning,
		}
		expectedDB := domain.Database{
			Plan: "postgresql-dr-enterprise-4096",
		}

		scClient.EXPECT().GetDatabase(ctx, databaseID).Return(currentDB, nil)
		scClient.EXPECT().ListFirewallRules(ctx, databaseID, addonID).Return(nil, nil)
		scClient.EXPECT().UpdateDatabasePlan(ctx, currentDB, expectedDB.Plan).
			Return(domain.DatabaseStatusProvisioning, nil)

		// When
		status, err := manager.UpdateDatabase(ctx, databaseID, expectedDB)

		// Then
		require.NoError(t, err)
		require.Equal(t, domain.DatabaseStatusProvisioning, status)
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
