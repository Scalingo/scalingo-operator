package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecret_String(t *testing.T) {
	t.Run("it redacts the value", func(t *testing.T) {
		secret := Secret{
			Namespace: "default",
			Name:      "scalingo",
			Key:       "api_token",
			Value:     "V3ry_s3cr3t_t0K3n",
		}
		require.Equal(t, "{ Namespace: default, Name: scalingo, Key: api_token, Value: [REDACTED] }", secret.String())
	})
}
