package scalingo

import (
	"context"

	scalingoapi "github.com/Scalingo/go-scalingo/v11"
	"github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func (c *client) GetDatabaseNetworkConfiguration(ctx context.Context, dbID string) (domain.DatabaseNetworkConfiguration, error) {
	config, err := c.scClient.Preview().DatabaseNetworkConfigurationShow(ctx, dbID)
	if err != nil {
		return domain.DatabaseNetworkConfiguration{}, errors.Wrap(ctx, err, "get database network configuration")
	}

	return domain.DatabaseNetworkConfiguration{
		OutscaleAccountID: config.OutscaleAccountID,
		OutscaleNetID:     config.OutscaleNetID,
		IPRange:           config.IPRange,
	}, nil
}

func (c *client) CreateDatabaseNetPeering(ctx context.Context, dbID, outscaleNetPeeringID string) (domain.DatabaseNetPeering, error) {
	netPeering, err := c.scClient.Preview().DatabaseNetPeeringCreate(ctx, dbID, scalingoapi.DatabaseNetPeeringCreateParams{
		OutscaleNetPeeringID: outscaleNetPeeringID,
	})
	if err != nil {
		return domain.DatabaseNetPeering{}, errors.Wrap(ctx, err, "create database net peering")
	}

	return domain.DatabaseNetPeering{
		ID:                   netPeering.ID,
		OutscaleNetPeeringID: netPeering.OutscaleNetPeeringID,
	}, nil
}

func (c *client) ListDatabaseNetPeerings(ctx context.Context, dbID string) ([]domain.DatabaseNetPeering, error) {
	netPeerings, err := c.scClient.Preview().DatabaseNetPeeringsList(ctx, dbID)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "list database net peerings")
	}

	result := make([]domain.DatabaseNetPeering, 0, len(netPeerings))
	for _, netPeering := range netPeerings {
		result = append(result, domain.DatabaseNetPeering{
			ID:                   netPeering.ID,
			OutscaleNetPeeringID: netPeering.OutscaleNetPeeringID,
		})
	}
	return result, nil
}

func (c *client) DeleteDatabaseNetPeering(ctx context.Context, dbID, netPeeringID string) error {
	err := c.scClient.Preview().DatabaseNetPeeringDestroy(ctx, dbID, netPeeringID)
	if err != nil {
		return errors.Wrap(ctx, err, "delete database net peering")
	}
	return nil
}
