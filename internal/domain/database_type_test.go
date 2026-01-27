package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseType_Validate(t *testing.T) {
	t.Run("it successfully validates type", func(t *testing.T) {
		require.NoError(t, DatabaseTypePostgreSQL.Validate())
	})

	t.Run("it returns error", func(t *testing.T) {
		require.ErrorContains(t, DatabaseType("").Validate(), "invalid database type")
		require.ErrorContains(t, DatabaseType("whatever").Validate(), "invalid database type")
	})
}
