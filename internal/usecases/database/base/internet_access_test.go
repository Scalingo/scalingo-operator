package database

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo/scalingomock"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestManager_updateInternetAccess(t *testing.T) {
	t.Run("it does nothing when internet access is already in expected state", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID: databaseID,
			Features: domain.DatabaseFeatures{
				domain.DatabaseFeatureForceSsl:          domain.DatabaseFeatureStatusActivated,
				domain.DatabaseFeaturePubliclyAvailable: domain.DatabaseFeatureStatusActivated,
			},
		}

		// When - expect enable and it's already enabled
		err := manager.updateInternetAccess(ctx, currentDB, true)

		// Then
		require.NoError(t, err)
	})

	t.Run("it does nothing when internet access disabled and expected disabled", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID:       databaseID,
			Features: domain.DatabaseFeatures{},
		}

		// When - expect disable and it's already disabled
		err := manager.updateInternetAccess(ctx, currentDB, false)

		// Then
		require.NoError(t, err)
	})

	t.Run("it enables force-ssl and publicly-available when enabling internet access", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID:       databaseID,
			Features: domain.DatabaseFeatures{},
		}

		scClient.EXPECT().EnableDatabaseFeature(gomock.Any(), currentDB, domain.DatabaseFeatureForceSsl).Return(nil)
		scClient.EXPECT().EnableDatabaseFeature(gomock.Any(), currentDB, domain.DatabaseFeaturePubliclyAvailable).Return(nil)

		// When
		err := manager.updateInternetAccess(ctx, currentDB, true)

		// Then
		require.NoError(t, err)
	})

	t.Run("it only enables publicly-available when force-ssl is already enabled", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID: databaseID,
			Features: domain.DatabaseFeatures{
				domain.DatabaseFeatureForceSsl: domain.DatabaseFeatureStatusActivated,
			},
		}

		scClient.EXPECT().EnableDatabaseFeature(gomock.Any(), currentDB, domain.DatabaseFeaturePubliclyAvailable).Return(nil)

		// When
		err := manager.updateInternetAccess(ctx, currentDB, true)

		// Then
		require.NoError(t, err)
	})

	t.Run("it fails when enabling force-ssl fails", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID:       databaseID,
			Features: domain.DatabaseFeatures{},
		}

		errEnableFeature := errors.New("enable feature force-ssl")
		scClient.EXPECT().EnableDatabaseFeature(gomock.Any(), currentDB, domain.DatabaseFeatureForceSsl).Return(errEnableFeature)

		// When
		err := manager.updateInternetAccess(ctx, currentDB, true)

		// Then
		require.ErrorIs(t, err, errEnableFeature)
	})

	t.Run("it fails when enabling publicly-available fails", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID: databaseID,
			Features: domain.DatabaseFeatures{
				domain.DatabaseFeatureForceSsl: domain.DatabaseFeatureStatusActivated,
			},
		}

		errEnableFeature := errors.New("enable feature publicly-available")
		scClient.EXPECT().EnableDatabaseFeature(gomock.Any(), currentDB, domain.DatabaseFeaturePubliclyAvailable).Return(errEnableFeature)

		// When
		err := manager.updateInternetAccess(ctx, currentDB, true)

		// Then
		require.ErrorIs(t, err, errEnableFeature)
	})

	t.Run("it disables publicly-available when disabling internet access", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID: databaseID,
			Features: domain.DatabaseFeatures{
				domain.DatabaseFeatureForceSsl:          domain.DatabaseFeatureStatusActivated,
				domain.DatabaseFeaturePubliclyAvailable: domain.DatabaseFeatureStatusActivated,
			},
		}

		scClient.EXPECT().DisableDatabaseFeature(gomock.Any(), currentDB, domain.DatabaseFeaturePubliclyAvailable).Return(nil)

		// When
		err := manager.updateInternetAccess(ctx, currentDB, false)

		// Then
		require.NoError(t, err)
	})

	t.Run("it fails when disabling publicly-available fails", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID: databaseID,
			Features: domain.DatabaseFeatures{
				domain.DatabaseFeatureForceSsl:          domain.DatabaseFeatureStatusActivated,
				domain.DatabaseFeaturePubliclyAvailable: domain.DatabaseFeatureStatusActivated,
			},
		}

		errDisableFeature := errors.New("disable feature publicly-available")
		scClient.EXPECT().DisableDatabaseFeature(gomock.Any(), currentDB, domain.DatabaseFeaturePubliclyAvailable).Return(errDisableFeature)

		// When
		err := manager.updateInternetAccess(ctx, currentDB, false)

		// Then
		require.ErrorIs(t, err, errDisableFeature)
	})
}

func TestExtractInternetAccessFeatures(t *testing.T) {
	t.Run("it returns false for both when features are empty", func(t *testing.T) {
		// Given
		currentDB := domain.Database{
			Features: domain.DatabaseFeatures{},
		}

		// When
		forceSsl, enable := extractInternetAccessFeatures(currentDB)

		// Then
		require.False(t, forceSsl)
		require.False(t, enable)
	})

	t.Run("it returns false for both when features are nil", func(t *testing.T) {
		// Given
		currentDB := domain.Database{
			Features: nil,
		}

		// When
		forceSsl, enable := extractInternetAccessFeatures(currentDB)

		// Then
		require.False(t, forceSsl)
		require.False(t, enable)
	})

	t.Run("it returns true for force-ssl when activated", func(t *testing.T) {
		// Given
		currentDB := domain.Database{
			Features: domain.DatabaseFeatures{
				domain.DatabaseFeatureForceSsl: domain.DatabaseFeatureStatusActivated,
			},
		}

		// When
		forceSsl, enable := extractInternetAccessFeatures(currentDB)

		// Then
		require.True(t, forceSsl)
		require.False(t, enable)
	})

	t.Run("it returns true for publicly-available when activated", func(t *testing.T) {
		// Given
		currentDB := domain.Database{
			Features: domain.DatabaseFeatures{
				domain.DatabaseFeaturePubliclyAvailable: domain.DatabaseFeatureStatusActivated,
			},
		}

		// When
		forceSsl, enable := extractInternetAccessFeatures(currentDB)

		// Then
		require.False(t, forceSsl)
		require.True(t, enable)
	})

	t.Run("it returns true for both when both activated", func(t *testing.T) {
		// Given
		currentDB := domain.Database{
			Features: domain.DatabaseFeatures{
				domain.DatabaseFeatureForceSsl:          domain.DatabaseFeatureStatusActivated,
				domain.DatabaseFeaturePubliclyAvailable: domain.DatabaseFeatureStatusActivated,
			},
		}

		// When
		forceSsl, enable := extractInternetAccessFeatures(currentDB)

		// Then
		require.True(t, forceSsl)
		require.True(t, enable)
	})

	t.Run("it returns true when features are pending", func(t *testing.T) {
		// Given
		currentDB := domain.Database{
			Features: domain.DatabaseFeatures{
				domain.DatabaseFeatureForceSsl:          domain.DatabaseFeatureStatusPending,
				domain.DatabaseFeaturePubliclyAvailable: domain.DatabaseFeatureStatusPending,
			},
		}

		// When
		forceSsl, enable := extractInternetAccessFeatures(currentDB)

		// Then
		require.True(t, forceSsl)
		require.True(t, enable)
	})

	t.Run("it returns false when features are failed", func(t *testing.T) {
		// Given
		currentDB := domain.Database{
			Features: domain.DatabaseFeatures{
				domain.DatabaseFeatureForceSsl:          domain.DatabaseFeatureStatusFailed,
				domain.DatabaseFeaturePubliclyAvailable: domain.DatabaseFeatureStatusFailed,
			},
		}

		// When
		forceSsl, enable := extractInternetAccessFeatures(currentDB)

		// Then
		require.False(t, forceSsl)
		require.False(t, enable)
	})
}
