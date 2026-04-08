package scalingo

import (
	"context"
	"fmt"
	"os"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	scalingoapi "github.com/Scalingo/go-scalingo/v10"
	errors "github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

type client struct {
	scClient *scalingoapi.Client
}

func NewClient(ctx context.Context, apiToken, region string) (scalingo.Client, error) {
	if apiToken == "" {
		return nil, errors.New(ctx, "empty api token")
	}

	log := logf.FromContext(ctx)

	userAgent := composeUserAgent(domain.Version)

	cfg := scalingoapi.ClientConfig{
		APIToken:  apiToken,
		Region:    region,
		UserAgent: userAgent,
	}

	scalingoAuthURL := os.Getenv("SCALINGO_AUTH_URL")
	if scalingoAuthURL != "" {
		log.Info("Set authentication end-point", "URL", scalingoAuthURL, "region", region)
		cfg.AuthEndpoint = scalingoAuthURL
	}

	scClient, err := scalingoapi.New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &client{
		scClient: scClient,
	}, nil
}

func composeUserAgent(version string) string {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	return fmt.Sprintf("%s %s", domain.AppName, version)
}
