package goscalingo

import (
	"context"
	"fmt"

	scalingo "github.com/Scalingo/go-scalingo/v8"
	errors "github.com/Scalingo/go-utils/errors/v2"

	goscalingo "github.com/Scalingo/scalingo-operator/internal/boundaries/out/go-scalingo"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

const (
	stagingRegion  = "osc-st-fr1"
	stagingAuthURL = "https://auth.st-sc.fr"

	postgresqlAddonProviderID = "postgresql-ng"
)

type client struct {
	scClient *scalingo.Client
}

func NewClient(ctx context.Context, apiToken, region string) (goscalingo.Client, error) {
	if apiToken == "" {
		return nil, errors.New(ctx, "empty token")
	}

	cfg := scalingo.ClientConfig{
		APIToken: apiToken,
		Region:   region,
	}

	// Ease execution on Staging.
	if region == stagingRegion {
		cfg.AuthEndpoint = stagingAuthURL
	}

	scClient, err := scalingo.New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &client{
		scClient: scClient,
	}, nil
}

func (c *client) CreateDatabase(ctx context.Context, db domain.Database) (domain.Database, error) {
	addonProviderID, err := toScalingoProviderId(db.Type)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "create database")
	}

	dbNG, err := c.scClient.Preview().DatabaseCreate(ctx, scalingo.DatabaseCreateParams{
		AddonProviderID: addonProviderID,
		PlanID:          db.Plan,
		Name:            db.Name,
		ProjectID:       db.ProjectID,
	})
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "create database")
	}

	return toDatabase(dbNG), nil
}

func (c *client) GetDatabase(ctx context.Context, dbID string) (domain.Database, error) {
	dbNG, err := c.scClient.Preview().DatabaseShow(ctx, dbID)
	if err != nil {
		return domain.Database{}, errors.Wrap(ctx, err, "get database")
	}

	return toDatabase(dbNG), nil
}

func (c *client) UpdateDatabase(ctx context.Context, db domain.Database) (domain.Database, error) {
	return domain.Database{}, domain.ErrNotImplemented
}

func (c *client) DeleteDatabase(ctx context.Context, dbID string) error {
	return domain.ErrNotImplemented
}

func (c *client) getConnectionURL(ctx context.Context, db domain.Database) (string, error) {
	return "", domain.ErrNotImplemented
}

func toScalingoProviderId(dbType domain.DatabaseType) (string, error) {
	switch dbType {
	case domain.DatabaseTypePostgreSQL:
		return postgresqlAddonProviderID, nil
	default:
		return "", fmt.Errorf("no matching provider for %q", dbType)
	}
}

func toAddonStatus(status scalingo.AddonStatus) domain.AddonStatus {
	switch status {
	case scalingo.AddonStatusProvisioning:
		return domain.AddonStatusProvisioning
	case scalingo.AddonStatusRunning:
		return domain.AddonStatusRunning
	default:
		return domain.AddonStatusSuspended
	}
}
func toDatabase(db scalingo.DatabaseNG) domain.Database {
	return domain.Database{
		ID:        db.App.ID,
		Name:      db.App.Name,
		Type:      domain.DatabaseType(db.Database.TypeName),
		Status:    toAddonStatus(db.Addon.Status),
		Plan:      db.Database.Plan,
		ProjectID: db.App.Project.ID,
	}
}
