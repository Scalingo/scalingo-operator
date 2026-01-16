package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestNewSecretManager(t *testing.T) {
	t.Run("creates a new SecretManager", func(t *testing.T) {
		manager := NewSecretManager(nil, nil)

		require.NotNil(t, manager)
	})
}

func TestSecretManager_GetSecret_Validation(t *testing.T) {
	ctx := t.Context()
	manager := NewSecretManager(nil, nil)

	t.Run("returns error when namespace is empty", func(t *testing.T) {
		secret := domain.Secret{
			Namespace: "",
			Name:      "test-secret",
			Key:       "api_token",
		}

		value, err := manager.GetSecret(ctx, secret)

		require.Error(t, err)
		require.Contains(t, err.Error(), "empty namespace")
		require.Empty(t, value)
	})

	t.Run("returns error when name is empty", func(t *testing.T) {
		secret := domain.Secret{
			Namespace: "default",
			Name:      "",
			Key:       "api_token",
		}

		value, err := manager.GetSecret(ctx, secret)

		require.Error(t, err)
		require.Contains(t, err.Error(), "empty name")
		require.Empty(t, value)
	})

	t.Run("returns error when key is empty", func(t *testing.T) {
		secret := domain.Secret{
			Namespace: "default",
			Name:      "test-secret",
			Key:       "",
		}

		value, err := manager.GetSecret(ctx, secret)

		require.Error(t, err)
		require.Contains(t, err.Error(), "empty key")
		require.Empty(t, value)
	})
}

func TestSecretManager_SetSecret_Validation(t *testing.T) {
	ctx := t.Context()
	manager := NewSecretManager(nil, nil)

	t.Run("returns error when namespace is empty", func(t *testing.T) {
		secret := domain.Secret{
			Namespace: "",
			Name:      "test-secret",
			Key:       "api_token",
			Value:     "token",
		}

		err := manager.SetSecret(ctx, secret)

		require.Error(t, err)
		require.Contains(t, err.Error(), "empty namespace")
	})

	t.Run("returns error when name is empty", func(t *testing.T) {
		secret := domain.Secret{
			Namespace: "default",
			Name:      "",
			Key:       "api_token",
			Value:     "token",
		}

		err := manager.SetSecret(ctx, secret)

		require.Error(t, err)
		require.Contains(t, err.Error(), "empty name")
	})

	t.Run("returns error when key is empty", func(t *testing.T) {
		secret := domain.Secret{
			Namespace: "default",
			Name:      "test-secret",
			Key:       "",
			Value:     "token",
		}

		err := manager.SetSecret(ctx, secret)

		require.Error(t, err)
		require.Contains(t, err.Error(), "empty key")
	})

	t.Run("returns error when value is empty", func(t *testing.T) {
		secret := domain.Secret{
			Namespace: "default",
			Name:      "test-secret",
			Key:       "api_token",
			Value:     "",
		}

		err := manager.SetSecret(ctx, secret)

		require.Error(t, err)
		require.Contains(t, err.Error(), "empty value")
	})
}
