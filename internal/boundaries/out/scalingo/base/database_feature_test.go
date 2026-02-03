package scalingo

import (
	"testing"

	"github.com/stretchr/testify/require"

	scalingoapi "github.com/Scalingo/go-scalingo/v9"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestToDatabaseFeatureStatus(t *testing.T) {
	tests := map[string]struct {
		scStatus        scalingoapi.DatabaseFeatureStatus
		expectedStatus  domain.DatabaseFeatureStatus
		isExpectedError bool
	}{
		"it converts activated status": {
			scStatus:       scalingoapi.DatabaseFeatureStatusActivated,
			expectedStatus: domain.DatabaseFeatureStatusActivated,
		},
		"it converts pending status": {
			scStatus:       scalingoapi.DatabaseFeatureStatusPending,
			expectedStatus: domain.DatabaseFeatureStatusPending,
		},
		"it converts failed status": {
			scStatus:       scalingoapi.DatabaseFeatureStatusFailed,
			expectedStatus: domain.DatabaseFeatureStatusFailed,
		},
		"it returns error for unknown status": {
			scStatus:        scalingoapi.DatabaseFeatureStatus("unknown"),
			isExpectedError: true,
		},
		"it returns error for empty status": {
			scStatus:        scalingoapi.DatabaseFeatureStatus(""),
			isExpectedError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			status, err := toDatabaseFeatureStatus(test.scStatus)

			if test.isExpectedError {
				require.Error(t, err)
				require.ErrorContains(t, err, "unknown database feature status")
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedStatus, status)
			}
		})
	}
}

func TestToDatabaseFeatures(t *testing.T) {
	ctx := t.Context()

	t.Run("it converts empty features list", func(t *testing.T) {
		scFeatures := []scalingoapi.DatabaseFeature{}

		features, err := toDatabaseFeatures(ctx, scFeatures)

		require.NoError(t, err)
		require.Empty(t, features)
	})

	t.Run("it converts single feature", func(t *testing.T) {
		scFeatures := []scalingoapi.DatabaseFeature{
			{
				Name:   "feature1",
				Status: scalingoapi.DatabaseFeatureStatusActivated,
			},
		}

		features, err := toDatabaseFeatures(ctx, scFeatures)

		require.NoError(t, err)
		require.Len(t, features, 1)
		require.Equal(t, domain.DatabaseFeatureStatusActivated, features["feature1"])
	})

	t.Run("it converts multiple features", func(t *testing.T) {
		scFeatures := []scalingoapi.DatabaseFeature{
			{
				Name:   "feature1",
				Status: scalingoapi.DatabaseFeatureStatusActivated,
			},
			{
				Name:   "feature2",
				Status: scalingoapi.DatabaseFeatureStatusPending,
			},
			{
				Name:   "feature3",
				Status: scalingoapi.DatabaseFeatureStatusFailed,
			},
		}

		features, err := toDatabaseFeatures(ctx, scFeatures)

		require.NoError(t, err)
		require.Len(t, features, 3)
		require.Equal(t, domain.DatabaseFeatureStatusActivated, features["feature1"])
		require.Equal(t, domain.DatabaseFeatureStatusPending, features["feature2"])
		require.Equal(t, domain.DatabaseFeatureStatusFailed, features["feature3"])
	})

	t.Run("it returns error for invalid feature status", func(t *testing.T) {
		scFeatures := []scalingoapi.DatabaseFeature{
			{
				Name:   "feature1",
				Status: scalingoapi.DatabaseFeatureStatus("invalid"),
			},
		}

		features, err := toDatabaseFeatures(ctx, scFeatures)

		require.Error(t, err)
		require.ErrorContains(t, err, "to database feature status")
		require.Empty(t, features)
	})

	t.Run("it returns error when encountering invalid status in multiple features", func(t *testing.T) {
		scFeatures := []scalingoapi.DatabaseFeature{
			{
				Name:   "feature1",
				Status: scalingoapi.DatabaseFeatureStatusActivated,
			},
			{
				Name:   "feature2",
				Status: scalingoapi.DatabaseFeatureStatus("invalid"),
			},
		}

		features, err := toDatabaseFeatures(ctx, scFeatures)

		require.Error(t, err)
		require.ErrorContains(t, err, "to database feature status")
		require.Empty(t, features)
	})
}
