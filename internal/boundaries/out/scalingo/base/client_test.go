package scalingo

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestNewClient(t *testing.T) {
	t.Run("it fails because of empty API token", func(t *testing.T) {
		ctx := t.Context()
		scClient, err := NewClient(ctx, "", stagingRegion)

		require.EqualError(t, err, "empty api token")
		require.Nil(t, scClient)
	})
}

func TestComposeUserAgent(t *testing.T) {
	t.Run("it prepends v when version does not start with it", func(t *testing.T) {
		userAgent := composeUserAgent("1.2.3")
		require.Equal(t, domain.AppName+" v1.2.3", userAgent)
	})

	t.Run("it keeps version unchanged when it already starts with v", func(t *testing.T) {
		userAgent := composeUserAgent("v1.2.3")
		require.Equal(t, domain.AppName+" v1.2.3", userAgent)
	})
}
