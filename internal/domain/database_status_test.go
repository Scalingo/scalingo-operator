package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseStatus_Validate(t *testing.T) {
	t.Run("it successfully validates status", func(t *testing.T) {
		require.NoError(t, DatabaseStatusRunning.Validate())
		require.NoError(t, DatabaseStatusProvisioning.Validate())
		require.NoError(t, DatabaseStatusSuspended.Validate())
	})

	t.Run("it returns error", func(t *testing.T) {
		require.ErrorContains(t, DatabaseStatus("").Validate(), "invalid database status")
		require.ErrorContains(t, DatabaseStatus("whatever").Validate(), "invalid database status")
	})
}
