package database

import (
	"context"

	"github.com/Scalingo/scalingo-operator/internal/domain"
)

type Manager interface {
	CreateDatabase(ctx context.Context, db domain.Database) (domain.Database, error)
	GetDatabase(ctx context.Context, dbID string) (domain.Database, error)
	GetDatabaseURL(ctx context.Context, db domain.Database) (domain.DatabaseURL, error)
	GetDatabaseEndpoints(ctx context.Context, dbID string) ([]domain.DatabaseEndpoint, error)
	GetDatabaseNetworkConfiguration(ctx context.Context, dbID string) (domain.DatabaseNetworkConfiguration, error)
	GetDatabaseNetPeerings(ctx context.Context, dbID string) ([]domain.DatabaseNetPeering, error)
	EnsureDatabaseNetPeering(ctx context.Context, dbID, outscaleNetPeeringID string) error
	DeleteDatabaseNetPeering(ctx context.Context, dbID, outscaleNetPeeringID string) error
	UpdateDatabase(ctx context.Context, dbID string, expectedDB domain.Database) (domain.DatabaseStatus, error)
	DeleteDatabase(ctx context.Context, dbID string) error
}
