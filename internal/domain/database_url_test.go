package domain

import (
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
