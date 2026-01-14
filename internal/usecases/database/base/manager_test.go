package database

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo/scalingomock"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestManager_CreateDatabase(t *testing.T) {
	t.Run("it successfully creates database", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		dbRequested := domain.Database{
			Name: "PG test",
			Type: domain.DatabaseTypePostgreSQL,
			Plan: "postgresql-ng-enterprise-4096",
		}

		dbCreated := dbRequested
		dbCreated.ID = "68f0ac70f4a35b9811b395a7"

		scClient.EXPECT().CreateDatabase(ctx, dbRequested).Return(dbCreated, nil)

		// When
		db, err := manager.CreateDatabase(ctx, dbRequested)

		// Then
		require.NoError(t, err)
		require.Equal(t, dbCreated, db)
	})
}
