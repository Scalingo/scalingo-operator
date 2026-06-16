package scalingo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	scalingoapi "github.com/Scalingo/go-scalingo/v11"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestClient_CreateDatabase(t *testing.T) {
	t.Run("it creates database with ip range", func(t *testing.T) {
		var payload map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v1/databases", r.URL.Path)

			decodeErr := json.NewDecoder(r.Body).Decode(&payload)
			assert.NoError(t, decodeErr)

			w.WriteHeader(http.StatusCreated)
			_, writeErr := w.Write([]byte(`{"database":{"id":"db-123","name":"my-db","app":{"id":"app-123"}}}`))
			assert.NoError(t, writeErr)
		}))
		defer server.Close()

		scClient, err := scalingoapi.New(t.Context(), scalingoapi.ClientConfig{
			APIEndpoint: server.URL,
			Region:      "",
		})
		require.NoError(t, err)

		client := client{scClient: scClient}
		_, err = client.CreateDatabase(t.Context(), domain.Database{
			Name:      "my-db",
			Type:      domain.DatabaseTypePostgreSQL,
			Plan:      "postgresql-dr-enterprise-4096",
			ProjectID: "prj-88888888-4444-4444-4444-cccccccccccc",
			IPRange:   "10.231.23.0/24",
		})

		require.NoError(t, err)
		assert.Equal(t, "10.231.23.0/24", payload["ip_range"])
	})

	t.Run("it omits empty ip range", func(t *testing.T) {
		var payload map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			decodeErr := json.NewDecoder(r.Body).Decode(&payload)
			assert.NoError(t, decodeErr)

			w.WriteHeader(http.StatusCreated)
			_, writeErr := w.Write([]byte(`{"database":{"id":"db-123","name":"my-db","app":{"id":"app-123"}}}`))
			assert.NoError(t, writeErr)
		}))
		defer server.Close()

		scClient, err := scalingoapi.New(t.Context(), scalingoapi.ClientConfig{
			APIEndpoint: server.URL,
			Region:      "",
		})
		require.NoError(t, err)

		client := client{scClient: scClient}
		_, err = client.CreateDatabase(t.Context(), domain.Database{
			Name: "my-db",
			Type: domain.DatabaseTypePostgreSQL,
			Plan: "postgresql-dr-enterprise-4096",
		})

		require.NoError(t, err)
		assert.NotContains(t, payload, "ip_range")
	})
}

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
