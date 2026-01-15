package scalingo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("it fails because of empty API token", func(t *testing.T) {
		ctx := t.Context()
		scClient, err := NewClient(ctx, "", stagingRegion)

		require.EqualError(t, err, "empty api token")
		require.Nil(t, scClient)
	})
}
