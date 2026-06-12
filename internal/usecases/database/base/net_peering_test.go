package database

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo/scalingomock"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestManager_GetDatabaseNetworkConfiguration(t *testing.T) {
	t.Run("it fails because of empty ID", func(t *testing.T) {
		ctx := t.Context()
		manager := manager{}
		res, err := manager.GetDatabaseNetworkConfiguration(ctx, "")

		require.EqualError(t, err, "empty database id")
		require.Empty(t, res)
	})

	t.Run("returns error when getting database network configuration fails", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		scClient.EXPECT().GetDatabaseNetworkConfiguration(ctx, databaseID).Return(domain.DatabaseNetworkConfiguration{}, errors.New("boom"))

		res, err := manager.GetDatabaseNetworkConfiguration(ctx, databaseID)

		require.EqualError(t, err, "get database network configuration: boom")
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

func TestManager_GetDatabaseNetPeerings(t *testing.T) {
	t.Run("fails because of empty ID", func(t *testing.T) {
		ctx := t.Context()
		manager := manager{}
		res, err := manager.GetDatabaseNetPeerings(ctx, "")

		require.EqualError(t, err, "empty database id")
		require.Empty(t, res)
	})

	t.Run("returns error when listing database net peerings fails", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		scClient.EXPECT().ListDatabaseNetPeerings(ctx, databaseID).Return(nil, errors.New("boom"))

		res, err := manager.GetDatabaseNetPeerings(ctx, databaseID)

		require.EqualError(t, err, "list database net peerings: boom")
		require.Empty(t, res)
	})

	t.Run("gets database net peerings", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		netPeerings := []domain.DatabaseNetPeering{
			{
				ID:                   "np-1",
				OutscaleNetPeeringID: "pcx-1234",
			},
			{
				ID:                   "np-2",
				OutscaleNetPeeringID: "pcx-5678",
			},
		}

		scClient.EXPECT().ListDatabaseNetPeerings(ctx, databaseID).Return(netPeerings, nil)

		res, err := manager.GetDatabaseNetPeerings(ctx, databaseID)

		require.NoError(t, err)
		require.Equal(t, netPeerings, res)
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

	t.Run("returns error when listing database endpoints fails", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		scClient.EXPECT().ListDatabaseEndpoints(ctx, databaseID).Return(nil, errors.New("boom"))

		res, err := manager.GetDatabaseEndpoints(ctx, databaseID)

		require.EqualError(t, err, "list database endpoints: boom")
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

	t.Run("returns error when listing database net peerings fails", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		scClient.EXPECT().ListDatabaseNetPeerings(ctx, databaseID).Return(nil, errors.New("boom"))

		err := manager.EnsureDatabaseNetPeering(ctx, databaseID, "pcx-1234")
		require.EqualError(t, err, "list database net peerings: boom")
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

	t.Run("returns error when creating database net peering fails", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		scClient.EXPECT().ListDatabaseNetPeerings(ctx, databaseID).Return([]domain.DatabaseNetPeering{}, nil)
		scClient.EXPECT().CreateDatabaseNetPeering(ctx, databaseID, "pcx-1234").Return(domain.DatabaseNetPeering{}, errors.New("boom"))

		err := manager.EnsureDatabaseNetPeering(ctx, databaseID, "pcx-1234")
		require.EqualError(t, err, "create database net peering: boom")
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

	t.Run("returns error when deleting database net peering fails", func(t *testing.T) {
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		scClient.EXPECT().DeleteDatabaseNetPeering(ctx, databaseID, "np-1").Return(errors.New("boom"))

		err := manager.DeleteDatabaseNetPeering(ctx, databaseID, "np-1")
		require.EqualError(t, err, "delete database net peering np-1: boom")
	})
}
