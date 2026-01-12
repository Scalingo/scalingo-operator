package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatusCondition_Validate(t *testing.T) {
	t.Run("it successfully validates status", func(t *testing.T) {
		require.NoError(t, DatabaseStatusConditionAvailable.Validate())
		require.NoError(t, DatabaseStatusConditionProvisioning.Validate())
	})

	t.Run("it returns error", func(t *testing.T) {
		require.ErrorContains(t, DatabaseStatusCondition("").Validate(), "invalid database status condition")
		require.ErrorContains(t, DatabaseStatusCondition("unknown").Validate(), "invalid database status condition")
	})
}
