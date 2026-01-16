package scalingo

import (
	"context"

	scalingoapi "github.com/Scalingo/go-scalingo/v9"
	errors "github.com/Scalingo/go-utils/errors/v3"

	"github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo"
)

const (
	stagingRegion  = "osc-st-fr1"
	stagingAuthURL = "https://auth.st-sc.fr"
)

type client struct {
	scClient *scalingoapi.Client
}

func NewClient(ctx context.Context, apiToken, region string) (scalingo.Client, error) {
	if apiToken == "" {
		return nil, errors.New(ctx, "empty api token")
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
