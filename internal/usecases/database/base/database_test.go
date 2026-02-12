package database

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo/scalingomock"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestManager_updateDatabasePlan(t *testing.T) {
	t.Run("it does nothing if current plan is the same as expected", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{Plan: "postgresql-dr-enterprise-4096"}
		expectedDB := domain.Database{Plan: "postgresql-dr-enterprise-4096"}

		err := manager.updateDatabasePlan(ctx, currentDB, expectedDB)
		require.NoError(t, err)
	})
}
