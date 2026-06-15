package domain

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseURL_String(t *testing.T) {
	t.Run("it redacts the value", func(t *testing.T) {
		dbURL := DatabaseURL{
			Name:  "SCALINGO_POSTGRESQL_URL",
			Value: "postgres://user:V3ry_s3cr3t_p4SSw0rd@host.fr:12345/db",
		}
		require.Equal(t, "{ Name: SCALINGO_POSTGRESQL_URL, Value: [REDACTED] }", dbURL.String())
	})
}

func TestComposeConnectionURLName(t *testing.T) {
	require.Equal(t, "SCALINGO_POSTGRESQL_URL", ComposeConnectionURLName("", "SCALINGO_POSTGRESQL_URL"))
	require.Equal(t, "PG_URL", ComposeConnectionURLName("PG", "SCALINGO_POSTGRESQL_URL"))
}

func TestComposeEndpointConnectionURLName(t *testing.T) {
	t.Run("uses default URL name as prefix when no prefix is set", func(t *testing.T) {
		name := ComposeEndpointConnectionURLName("", "SCALINGO_POSTGRESQL_URL", DatabaseEndpointTypePublicRW)

		require.Equal(t, "SCALINGO_POSTGRESQL_PUBLIC_RW_URL", name)
	})

	t.Run("uses the configured prefix when set", func(t *testing.T) {
		name := ComposeEndpointConnectionURLName("PG", "SCALINGO_POSTGRESQL_URL", DatabaseEndpointTypePrivatePeeringRW)

		require.Equal(t, "PG_PRIVATE_PEERING_RW_URL", name)
	})
}

func TestComposeEndpointConnectionURL(t *testing.T) {
	t.Run("replaces host and port from the endpoint", func(t *testing.T) {
		endpointURL, err := ComposeEndpointConnectionURL(t.Context(), "postgres://user:password@original-host:1234/db?sslmode=require", DatabaseEndpoint{
			Hostname: "endpoint-host",
			Port:     5432,
		})

		require.NoError(t, err)
		require.Equal(t, "postgres://user:password@endpoint-host:5432/db?sslmode=require", endpointURL)
	})

	t.Run("escapes endpoint host safely", func(t *testing.T) {
		endpointURL, err := ComposeEndpointConnectionURL(t.Context(), "postgres://user:password@original-host:1234/db", DatabaseEndpoint{
			Hostname: "2001:db8::1",
			Port:     5432,
		})

		require.NoError(t, err)
		parsedURL, err := url.Parse(endpointURL)
		require.NoError(t, err)
		require.Equal(t, "[2001:db8::1]:5432", parsedURL.Host)
	})
}
