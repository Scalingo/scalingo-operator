package scalingo

import (
	"context"
	"fmt"

	scalingoapi "github.com/Scalingo/go-scalingo/v8"
	errors "github.com/Scalingo/go-utils/errors/v2"

	scalingo "github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

const (
	stagingRegion  = "osc-st-fr1"
	stagingAuthURL = "https://auth.st-sc.fr"

	postgresqlAddonProviderID = "postgresql-ng"
)

type client struct {
	scClient *scalingoapi.Client
}

func NewClient(ctx context.Context, apiToken, region string) (scalingo.Client, error) {
	if apiToken == "" {
		return nil, errors.New(ctx, "empty token")
	}

	cfg := scalingoapi.ClientConfig{
		APIToken: apiToken,
		Region:   region,
	}

	// Ease execution on Staging.
	if region == stagingRegion {
		cfg.AuthEndpoint = stagingAuthURL
	}

	scClient, err := scalingoapi.New(ctx, cfg)
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

	dbNG, err := c.scClient.Preview().DatabaseCreate(ctx, scalingoapi.DatabaseCreateParams{
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

func toScalingoProviderId(dbType domain.DatabaseType) (string, error) {
	switch dbType {
	case domain.DatabaseTypePostgreSQL:
		return postgresqlAddonProviderID, nil
	default:
		return "", fmt.Errorf("no matching provider for %q", dbType)
	}
}

func toAddonStatus(status scalingoapi.AddonStatus) domain.AddonStatus {
	switch status {
	case scalingoapi.AddonStatusProvisioning:
		return domain.AddonStatusProvisioning
	case scalingoapi.AddonStatusRunning:
		return domain.AddonStatusRunning
	default:
		return domain.AddonStatusSuspended
	}
}
func toDatabase(db scalingoapi.DatabaseNG) domain.Database {
	return domain.Database{
		ID:        db.App.ID,
		Name:      db.App.Name,
		Type:      domain.DatabaseType(db.Database.TypeName),
		Status:    toAddonStatus(db.Addon.Status),
		Plan:      db.Database.Plan,
		ProjectID: db.App.Project.ID,
	}
}
