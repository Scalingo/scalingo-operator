package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFinalizerConstants(t *testing.T) {
	t.Run("PostgreSQLFinalizerName has correct value", func(t *testing.T) {
		require.Equal(t, "databases.scalingo.com/PostgresFinalizer", PostgreSQLFinalizerName)
	})
}
