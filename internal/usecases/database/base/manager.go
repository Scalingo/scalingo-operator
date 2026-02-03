package database

import (
	"context"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

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
		return nil, errors.Wrap(ctx, err, "new manager")
	}
	if apiToken == "" {
		return nil, errors.New(ctx, "empty api token")
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
	log := logf.FromContext(ctx)

	if dbID == "" {
		return domain.Database{}, errors.New(ctx, "empty database id")
	}
	db, err := m.scClient.GetDatabase(ctx, dbID)
	if err != nil {
		return domain.Database{}, errors.Wrapf(ctx, err, "get database %s", dbID)
	}

	rules, err := m.scClient.ListFirewallRules(ctx, db.ID, db.AddonID)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "get database firewall rules")
	}

	db.FireWallRules = rules
	log.Info("Get database", "database", db)

	return db, nil
}

func (m *manager) GetDatabaseURL(ctx context.Context, db domain.Database) (domain.DatabaseURL, error) {
	dbTypeName, err := toDatabaseTypeName(ctx, db.Type)
	if err != nil {
		return domain.DatabaseURL{}, errors.Wrap(ctx, err, "to database type name")
	}

	varName := "SCALINGO_" + dbTypeName + "_URL"
	varValue, err := m.scClient.FindApplicationVariable(ctx, db.AppID, varName)
	if err != nil {
		return domain.DatabaseURL{}, errors.Wrap(ctx, err, "find database url")
	}
	return domain.DatabaseURL{
		Name:  varName,
		Value: varValue,
	}, nil
}

func (m *manager) UpdateDatabase(ctx context.Context, dbID string, expectedDB domain.Database) error {
	currentDB, err := m.GetDatabase(ctx, dbID)
	if err != nil {
		return errors.Wrapf(ctx, err, "unreachable database %s", dbID)
	}

	// Internet public access.
	publiclyAvailable, ok := expectedDB.Features[domain.DatabaseFeaturePubliclyAvailable]
	expectedEnableInternetAccess := ok && publiclyAvailable == domain.DatabaseFeatureStatusActivated

	err = m.updateInternetAccess(ctx, currentDB, expectedEnableInternetAccess)
	if err != nil {
		return errors.Wrap(ctx, err, "update internet access")
	}

	// Firewall rules.
	err = m.updateFirewallRules(ctx, currentDB, expectedDB.FireWallRules)
	if err != nil {
		return errors.Wrap(ctx, err, "update firewall rules")
	}

	return nil
}

func (m *manager) DeleteDatabase(ctx context.Context, dbID string) error {
	if dbID == "" {
		return errors.New(ctx, "empty database id")
	}

	err := m.scClient.DeleteDatabase(ctx, dbID)
	if err != nil {
		return errors.Wrapf(ctx, err, "delete database %v", dbID)
	}
	return nil
}

func toDatabaseTypeName(ctx context.Context, dbType domain.DatabaseType) (string, error) {
	switch dbType {
	case domain.DatabaseTypePostgreSQL:
		return "POSTGRESQL", nil
	default:
		return "", errors.Newf(ctx, "no matching type for %q", dbType)
	}
}
