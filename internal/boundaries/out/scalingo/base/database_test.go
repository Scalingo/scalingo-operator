package scalingo

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestClient_UpdateDatabasePlan(t *testing.T) {
	t.Run("it fails if current plan is the same as expected", func(t *testing.T) {
		const dbPlan = "postgresql-dr-enterprise-4096"

		currentDB := domain.Database{
			Plan: dbPlan,
		}
		expectedPlan := dbPlan

		scClient := client{}
		_, err := scClient.UpdateDatabasePlan(t.Context(), currentDB, expectedPlan)
		require.ErrorIs(t, err, domain.ErrNothingToBeDone)
	})
}
