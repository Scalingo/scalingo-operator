package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseFeatureStatus_Validate(t *testing.T) {
	t.Run("it successfully validates status", func(t *testing.T) {
		require.NoError(t, DatabaseFeatureStatusActivated.Validate())
		require.NoError(t, DatabaseFeatureStatusPending.Validate())
		require.NoError(t, DatabaseFeatureStatusFailed.Validate())
	})

	t.Run("it returns error for unknown status", func(t *testing.T) {
		require.ErrorContains(t, DatabaseFeatureStatus("").Validate(), "invalid database feature status")
		require.ErrorContains(t, DatabaseFeatureStatus("unknown").Validate(), "invalid database feature status")
	})
}

func TestDatabaseFeatureStatus_IsActive(t *testing.T) {
	t.Run("it returns true for activated status", func(t *testing.T) {
		require.True(t, DatabaseFeatureStatusActivated.IsActive())
	})

	t.Run("it returns true for pending status", func(t *testing.T) {
		require.True(t, DatabaseFeatureStatusPending.IsActive())
	})

	t.Run("it returns false for failed status", func(t *testing.T) {
		require.False(t, DatabaseFeatureStatusFailed.IsActive())
	})

	t.Run("it returns false for unknown status", func(t *testing.T) {
		require.False(t, DatabaseFeatureStatus("").IsActive())
		require.False(t, DatabaseFeatureStatus("unknown").IsActive())
	})
}
