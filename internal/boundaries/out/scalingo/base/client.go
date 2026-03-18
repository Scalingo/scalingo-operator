package scalingo

import (
	"context"
	"fmt"

	scalingoapi "github.com/Scalingo/go-scalingo/v10"
	errors "github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

const (
	stagingRegion  = "osc-st-fr1"
	stagingAuthURL = "https://auth.st-sc.fr"

	localRegion  = "local"
	localAuthURL = "http://172.17.0.1:1234"
)

type client struct {
	scClient *scalingoapi.Client
}

func NewClient(ctx context.Context, apiToken, region string) (scalingo.Client, error) {
	if apiToken == "" {
		return nil, errors.New(ctx, "empty api token")
	}

	userAgent := fmt.Sprintf("%s v%s", domain.AppName, domain.Version)

	cfg := scalingoapi.ClientConfig{
		APIToken:  apiToken,
		Region:    region,
		UserAgent: userAgent,
	}

	// Auth endpoints for Staging and Local environments.
	switch region {
	case stagingRegion:
		cfg.AuthEndpoint = stagingAuthURL
	case localRegion:
		cfg.AuthEndpoint = localAuthURL
	}

	scClient, err := scalingoapi.New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &client{
		scClient: scClient,
	}, nil
}
