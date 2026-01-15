package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsDatabaseRunning(t *testing.T) {
	t.Run("returns true when annotation exists and is true", func(t *testing.T) {
		dbMeta := metav1.ObjectMeta{
			Annotations: map[string]string{
				DatabaseAnnotationIsRunning: "true",
			},
		}
		require.True(t, IsDatabaseRunning(dbMeta))
	})

	t.Run("returns false when annotation exists but is false", func(t *testing.T) {
		dbMeta := metav1.ObjectMeta{
			Annotations: map[string]string{
				DatabaseAnnotationIsRunning: "false",
			},
		}
		require.False(t, IsDatabaseRunning(dbMeta))
	})

	t.Run("returns false when annotation does not exist", func(t *testing.T) {
		dbMeta := metav1.ObjectMeta{
			Annotations: map[string]string{},
		}
		require.False(t, IsDatabaseRunning(dbMeta))
	})

	t.Run("returns false when annotations is nil", func(t *testing.T) {
		dbMeta := metav1.ObjectMeta{}
		require.False(t, IsDatabaseRunning(dbMeta))
	})
}

func TestIsDatabaseDeletionRequested(t *testing.T) {
	t.Run("returns true when deletion timestamp is set", func(t *testing.T) {
		now := metav1.Now()
		dbMeta := metav1.ObjectMeta{
			DeletionTimestamp: &now,
		}
		require.True(t, IsDatabaseDeletionRequested(dbMeta))
	})

	t.Run("returns false when deletion timestamp is not set", func(t *testing.T) {
		dbMeta := metav1.ObjectMeta{}
		require.False(t, IsDatabaseDeletionRequested(dbMeta))
	})
}

func TestIsDatabaseAvailable(t *testing.T) {
	t.Run("returns true when Available condition is true", func(t *testing.T) {
		conditions := []metav1.Condition{
			{
				Type:   string(DatabaseStatusConditionAvailable),
				Status: metav1.ConditionTrue,
			},
		}
		require.True(t, IsDatabaseAvailable(conditions))
	})

	t.Run("returns false when Available condition is false", func(t *testing.T) {
		conditions := []metav1.Condition{
			{
				Type:   string(DatabaseStatusConditionAvailable),
				Status: metav1.ConditionFalse,
			},
		}
		require.False(t, IsDatabaseAvailable(conditions))
	})

	t.Run("returns false when Available condition does not exist", func(t *testing.T) {
		conditions := []metav1.Condition{}
		require.False(t, IsDatabaseAvailable(conditions))
	})
}

func TestIsDatabaseProvisioning(t *testing.T) {
	t.Run("returns true when Provisioning condition is true", func(t *testing.T) {
		conditions := []metav1.Condition{
			{
				Type:   string(DatabaseStatusConditionProvisioning),
				Status: metav1.ConditionTrue,
			},
		}
		require.True(t, IsDatabaseProvisioning(conditions))
	})

	t.Run("returns false when Provisioning condition is false", func(t *testing.T) {
		conditions := []metav1.Condition{
			{
				Type:   string(DatabaseStatusConditionProvisioning),
				Status: metav1.ConditionFalse,
			},
		}
		require.False(t, IsDatabaseProvisioning(conditions))
	})

	t.Run("returns false when Provisioning condition does not exist", func(t *testing.T) {
		conditions := []metav1.Condition{}
		require.False(t, IsDatabaseProvisioning(conditions))
	})
}

func TestSetDatabaseInitialState(t *testing.T) {
	t.Run("sets initial state correctly", func(t *testing.T) {
		dbMeta := &metav1.ObjectMeta{}
		conditions := &[]metav1.Condition{}

		SetDatabaseInitialState(dbMeta, conditions)

		require.Equal(t, "false", dbMeta.Annotations[DatabaseAnnotationIsRunning])
		require.Len(t, *conditions, 2)
		require.False(t, IsDatabaseAvailable(*conditions))
		require.False(t, IsDatabaseProvisioning(*conditions))
	})

	t.Run("sets correct reasons and messages", func(t *testing.T) {
		dbMeta := &metav1.ObjectMeta{}
		conditions := &[]metav1.Condition{}

		SetDatabaseInitialState(dbMeta, conditions)

		require.Equal(t, reasonNotAvailable, (*conditions)[0].Reason)
		require.Equal(t, msgNotAvailable, (*conditions)[0].Message)
		require.Equal(t, reasonNotprovisioned, (*conditions)[1].Reason)
		require.Equal(t, msgNotProvisioned, (*conditions)[1].Message)
	})
}

func TestSetDatabaseStatusProvisioning(t *testing.T) {
	t.Run("sets provisioning status correctly", func(t *testing.T) {
		conditions := &[]metav1.Condition{}

		SetDatabaseStatusProvisioning(conditions)

		require.Len(t, *conditions, 1)
		require.True(t, IsDatabaseProvisioning(*conditions))
		require.Equal(t, reasonProvisioning, (*conditions)[0].Reason)
		require.Equal(t, msgProvisioning, (*conditions)[0].Message)
	})

	t.Run("updates existing provisioning condition", func(t *testing.T) {
		conditions := &[]metav1.Condition{
			{
				Type:   string(DatabaseStatusConditionProvisioning),
				Status: metav1.ConditionFalse,
				Reason: "OldReason",
			},
		}

		SetDatabaseStatusProvisioning(conditions)

		require.Len(t, *conditions, 1)
		require.True(t, IsDatabaseProvisioning(*conditions))
		require.Equal(t, reasonProvisioning, (*conditions)[0].Reason)
	})
}

func TestSetDatabaseStatusProvisioned(t *testing.T) {
	t.Run("sets provisioned status correctly", func(t *testing.T) {
		dbMeta := &metav1.ObjectMeta{}
		conditions := &[]metav1.Condition{}

		SetDatabaseStatusProvisioned(dbMeta, conditions)

		require.Equal(t, "true", dbMeta.Annotations[DatabaseAnnotationIsRunning])
		require.Len(t, *conditions, 2)
		require.True(t, IsDatabaseAvailable(*conditions))
		require.False(t, IsDatabaseProvisioning(*conditions))
	})

	t.Run("sets correct reasons and messages", func(t *testing.T) {
		dbMeta := &metav1.ObjectMeta{}
		conditions := &[]metav1.Condition{}

		SetDatabaseStatusProvisioned(dbMeta, conditions)

		require.Len(t, *conditions, 2)
		for _, cond := range *conditions {
			if cond.Type == string(DatabaseStatusConditionAvailable) {
				require.Equal(t, reasonAvailable, cond.Reason)
				require.Equal(t, msgAvailable, cond.Message)
			} else {
				require.Equal(t, string(DatabaseStatusConditionProvisioning), cond.Type)
				require.Equal(t, reasonProvisioned, cond.Reason)
				require.Equal(t, msgProvisioned, cond.Message)
			}
		}
	})
}
