package database

import (
	"context"

	errors "github.com/Scalingo/go-utils/errors/v3"

	scalingo "github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo"
	scalingobase "github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo/base"
	"github.com/Scalingo/scalingo-operator/internal/domain"
	"github.com/Scalingo/scalingo-operator/internal/usecases/database"
)

type manager struct {
	dbType   domain.DatabaseType
	scClient scalingo.Client
}

func NewManager(ctx context.Context, dbType domain.DatabaseType, apiToken, region string) (database.Manager, error) {
	err := dbType.Validate()
	if err != nil {
		return nil, errors.Newf(ctx, "new manager: %v", err)
	}
	if apiToken == "" {
		return nil, errors.New(ctx, "empty apitoken")
	}

	scClient, err := scalingobase.NewClient(ctx, apiToken, region)
	if err != nil {
		return nil, errors.Wrap(ctx, err, "new scalingo client")
	}

	return &manager{
		dbType:   dbType,
		scClient: scClient,
	}, nil
}

func (m *manager) CreateDatabase(ctx context.Context, db domain.Database) (domain.Database, error) {
	return m.scClient.CreateDatabase(ctx, db)
}

func (m *manager) GetDatabase(ctx context.Context, dbID string) (domain.Database, error) {
	if dbID == "" {
		return domain.Database{}, errors.New(ctx, "empty database id")
	}
	return m.scClient.GetDatabase(ctx, dbID)
}

func (m *manager) UpdateDatabase(ctx context.Context, currentDB, expectedDB domain.Database) (domain.Database, error) {
	return domain.Database{}, domain.ErrNotImplemented
}
func (m *manager) DeleteDatabase(ctx context.Context, dbID string) error {
	if dbID == "" {
		return errors.New(ctx, "empty database id")
	}
	return m.scClient.DeleteDatabase(ctx, dbID)
}
