package apiclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoJSONSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1/resource", r.URL.Path)
		require.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var in map[string]string
		require.NoError(t, json.Unmarshal(body, &in))
		require.Equal(t, "value", in["key"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	client := New(server.URL+"/", server.Client())
	out := struct {
		OK bool `json:"ok"`
	}{}

	err := client.DoJSON(
		context.Background(),
		http.MethodPost,
		"/v1/resource",
		map[string]string{"Authorization": "Bearer token"},
		map[string]string{"key": "value"},
		&out,
	)

	require.NoError(t, err)
	assert.True(t, out.OK)
}

func TestDoJSONReturnsHTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("  unauthorized  \n"))
	}))
	t.Cleanup(server.Close)

	client := New(server.URL, server.Client())
	err := client.DoJSON(context.Background(), http.MethodGet, "/v1/protected", nil, nil, nil)
	require.Error(t, err)

	httpErr, ok := err.(*HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.StatusCode)
	assert.Equal(t, "unauthorized", httpErr.Body)
}

func TestDoJSONReturnsNilForEmptyBodyWhenOutNil(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	client := New(server.URL, server.Client())
	err := client.DoJSON(context.Background(), http.MethodDelete, "/v1/resource", nil, nil, nil)
	assert.NoError(t, err)
}
