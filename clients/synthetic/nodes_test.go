// @license
// Copyright 2026 Dynatrace LLC
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package synthetic_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/synthetic"
)

func TestListNodes(t *testing.T) {
	const response = `{"nodes":[{"entityId":"SYNTHETIC_ENGINE-1"},{"entityId":"SYNTHETIC_ENGINE-2"}]}`
	yes, no := true, false

	t.Run("extracts nodes and sends filters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/nodes", r.URL.Path)
			require.Equal(t, "true", r.URL.Query().Get("assigned-to-location"))
			require.Equal(t, "false", r.URL.Query().Get("containerized"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.ListNodes(t.Context(), synthetic.ListNodesOptions{AssignedToLocation: &yes, Containerized: &no})

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, resp.Objects, 2)
		assert.JSONEq(t, `{"entityId":"SYNTHETIC_ENGINE-1"}`, string(resp.Objects[0]))
	})

	t.Run("omits filters when nil", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/nodes", r.URL.Path)
			require.Empty(t, r.URL.RawQuery)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.ListNodes(t.Context(), synthetic.ListNodesOptions{})

		require.NoError(t, err)
	})

	t.Run("sends only set filters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/nodes", r.URL.Path)
			require.Equal(t, "true", r.URL.Query().Get("containerized"))
			require.NotContains(t, r.URL.Query(), "assigned-to-location")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.ListNodes(t.Context(), synthetic.ListNodesOptions{Containerized: &yes})

		require.NoError(t, err)
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/nodes", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.ListNodes(t.Context(), synthetic.ListNodesOptions{})

		assert.ErrorAs(t, err, new(api.APIError))
		assert.Empty(t, resp)
	})

	t.Run("invalid JSON returns RuntimeError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/nodes", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not json"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.ListNodes(t.Context(), synthetic.ListNodesOptions{})

		var rtErr api.RuntimeError
		require.ErrorAs(t, err, &rtErr)
		assert.Equal(t, "synthetic nodes", rtErr.Resource)
		assert.Empty(t, resp)
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.ListNodes(t.Context(), synthetic.ListNodesOptions{})

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic nodes", clientErr.Resource)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Empty(t, resp)
	})
}

func TestGetNode(t *testing.T) {
	const response = `{"entityId":"SYNTHETIC_ENGINE-1"}`

	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/nodes/SYNTHETIC_ENGINE-1", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.GetNode(t.Context(), "SYNTHETIC_ENGINE-1")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, response, string(resp.Data))
	})

	t.Run("empty id returns error and sends no request", func(t *testing.T) {
		client := synthetic.NewClient(&rest.Client{})

		resp, err := client.GetNode(t.Context(), "")

		assert.ErrorIs(t, err, api.ValidationError{Resource: "synthetic nodes", Field: "id", Reason: "is empty"})
		assert.Empty(t, resp)
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/nodes/id", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.GetNode(t.Context(), "id")

		assert.Empty(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "boom", string(apiErr.Body))
	})

	t.Run("404 is a not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/nodes/id", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.GetNode(t.Context(), "id")

		assert.True(t, api.IsNotFoundError(err))
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.GetNode(t.Context(), "id")

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic nodes", clientErr.Resource)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Empty(t, resp)
	})
}
