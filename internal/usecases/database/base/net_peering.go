package database

import (
	"context"

	"github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func (m *manager) GetDatabaseNetworkConfiguration(ctx context.Context, dbID string) (domain.DatabaseNetworkConfiguration, error) {
	if dbID == "" {
		return domain.DatabaseNetworkConfiguration{}, errors.New(ctx, "empty database id")
	}

	config, err := m.scClient.GetDatabaseNetworkConfiguration(ctx, dbID)
	if err != nil {
		return domain.DatabaseNetworkConfiguration{}, errors.Wrap(ctx, err, "get database network configuration")
	}

	return config, nil
}

func (m *manager) GetDatabaseNetPeerings(ctx context.Context, dbID string) ([]domain.DatabaseNetPeering, error) {
	if dbID == "" {
		return nil, errors.New(ctx, "empty database id")
	}

	netPeerings, err := m.scClient.ListDatabaseNetPeerings(ctx, dbID)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "list database net peerings")
	}

	return netPeerings, nil
}

func (m *manager) GetDatabaseEndpoints(ctx context.Context, dbID string) ([]domain.DatabaseEndpoint, error) {
	if dbID == "" {
		return nil, errors.New(ctx, "empty database id")
	}

	endpoints, err := m.scClient.ListDatabaseEndpoints(ctx, dbID)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "list database endpoints")
	}

	return endpoints, nil
}

func (m *manager) EnsureDatabaseNetPeering(ctx context.Context, dbID, outscaleNetPeeringID string) error {
	if dbID == "" {
		return errors.New(ctx, "empty database id")
	}
	if outscaleNetPeeringID == "" {
		return errors.New(ctx, "empty outscale net peering id")
	}

	netPeerings, err := m.scClient.ListDatabaseNetPeerings(ctx, dbID)
	if err != nil {
		return errors.Wrap(ctx, err, "list database net peerings")
	}

	for _, netPeering := range netPeerings {
		if netPeering.OutscaleNetPeeringID == outscaleNetPeeringID {
			return nil
		}
	}

	_, err = m.scClient.CreateDatabaseNetPeering(ctx, dbID, outscaleNetPeeringID)
	if err != nil {
		return errors.Wrap(ctx, err, "create database net peering")
	}
	return nil
}

func (m *manager) DeleteDatabaseNetPeering(ctx context.Context, dbID string, netPeeringID string) error {
	if dbID == "" {
		return errors.New(ctx, "empty database id")
	}
	if netPeeringID == "" {
		return errors.New(ctx, "empty net peering id")
	}

	err := m.scClient.DeleteDatabaseNetPeering(ctx, dbID, netPeeringID)
	if err != nil {
		return errors.Wrapf(ctx, err, "delete database net peering %s", netPeeringID)
	}
	return nil
}
