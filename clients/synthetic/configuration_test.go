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
	"io"
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

func TestGetConfiguration(t *testing.T) {
	const response = `{"someSetting":true}`

	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/configuration", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.GetConfiguration(t.Context())
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, response, string(resp.Data))
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/configuration", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.GetConfiguration(t.Context())

		assert.Empty(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "boom", string(apiErr.Body))
	})

	t.Run("404 is a not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/configuration", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.GetConfiguration(t.Context())

		assert.True(t, api.IsNotFoundError(err))
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.GetConfiguration(t.Context())

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic configuration", clientErr.Resource)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Empty(t, resp)
	})
}

func TestUpdateConfiguration(t *testing.T) {
	const body = `{"someSetting":false}`

	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/synthetic/v1/configuration", r.URL.Path)
			b, _ := io.ReadAll(r.Body)
			require.Equal(t, body, string(b))
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.UpdateConfiguration(t.Context(), []byte(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/synthetic/v1/configuration", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.UpdateConfiguration(t.Context(), []byte(body))

		assert.Empty(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "boom", string(apiErr.Body))
	})

	t.Run("404 is a not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/synthetic/v1/configuration", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.UpdateConfiguration(t.Context(), []byte(body))

		assert.True(t, api.IsNotFoundError(err))
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.UpdateConfiguration(t.Context(), []byte(body))

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic configuration", clientErr.Resource)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Empty(t, resp)
	})
}
