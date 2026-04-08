package scalingo

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestNewClient(t *testing.T) {
	t.Run("it fails because of empty API token", func(t *testing.T) {
		ctx := t.Context()
		scClient, err := NewClient(ctx, "", "osc-fr1")

		require.EqualError(t, err, "empty api token")
		require.Nil(t, scClient)
	})

	t.Run("it builds an authenticated client when region lookup is not needed", func(t *testing.T) {
		ctx := t.Context()

		scClient, err := NewClient(ctx, "token", "")

		require.NoError(t, err)
		require.NotNil(t, scClient)

		baseClient, ok := scClient.(*client)
		require.True(t, ok)
		require.NotNil(t, baseClient.scClient)
		require.True(t, baseClient.scClient.AuthAPI().IsAuthenticatedClient())
		require.True(t, baseClient.scClient.ScalingoAPI().IsAuthenticatedClient())
	})

	t.Run("it uses the auth endpoint from the environment when provided", func(t *testing.T) {
		t.Setenv("SCALINGO_AUTH_URL", "https://auth.example.test")

		ctx := t.Context()
		scClient, err := NewClient(ctx, "token", "")

		require.NoError(t, err)

		baseClient, ok := scClient.(*client)
		require.True(t, ok)
		require.Equal(t, "https://auth.example.test/v1", baseClient.scClient.AuthAPI().BaseURL())
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
