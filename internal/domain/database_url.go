package domain

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/Scalingo/go-utils/errors/v3"
)

const ConnectionURLNameSuffix = "_URL"

type DatabaseURL struct {
	Name  string
	Value string // WARNING: contains password.
}

func (u DatabaseURL) String() string {
	return fmt.Sprintf("{ Name: %s, Value: %s }", u.Name, Redacted)
}

func ComposeConnectionURLName(prefix, defaultName string) string {
	if prefix == "" {
		return defaultName
	}
	return prefix + ConnectionURLNameSuffix
}

func ComposeEndpointConnectionURLName(prefix, defaultName string, endpointType DatabaseEndpointType) string {
	if prefix == "" {
		prefix = strings.TrimSuffix(defaultName, ConnectionURLNameSuffix)
	}
	return prefix + "_" + strings.ToUpper(strings.ReplaceAll(string(endpointType), "-", "_")) + ConnectionURLNameSuffix
}

func ComposeEndpointConnectionURL(ctx context.Context, defaultURL string, endpoint DatabaseEndpoint) (string, error) {
	parsedURL, err := url.Parse(defaultURL)
	if err != nil {
		return "", errors.Wrap(ctx, err, "parse database url")
	}
	parsedURL.Host = net.JoinHostPort(endpoint.Hostname, strconv.Itoa(endpoint.Port))
	return parsedURL.String(), nil
}
