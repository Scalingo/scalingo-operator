package helpers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRequeueDelay(t *testing.T) {
	t.Run("RequeueDelay has correct value", func(t *testing.T) {
		require.Equal(t, 30*time.Second, RequeueDelay)
	})
}
