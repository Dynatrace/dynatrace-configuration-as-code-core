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

func TestListMonitors(t *testing.T) {
	const response = `{"monitors":[{"id":"HTTP_CHECK-1","name":"a","type":"HTTP","enabled":true},{"id":"HTTP_CHECK-2","name":"b","type":"HTTP","enabled":false}]}`

	t.Run("extracts monitors and sends selector", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors", r.URL.Path)
			require.Equal(t, "externalId(abc),enabled(true)", r.URL.Query().Get("monitor-selector"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.ListMonitors(t.Context(), synthetic.ListMonitorsOptions{MonitorSelector: "externalId(abc),enabled(true)"})

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, resp.Objects, 2)
		assert.JSONEq(t, `{"id":"HTTP_CHECK-1","name":"a","type":"HTTP","enabled":true}`, string(resp.Objects[0]))
	})

	t.Run("omits selector when empty", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors", r.URL.Path)
			require.Empty(t, r.URL.RawQuery)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.ListMonitors(t.Context(), synthetic.ListMonitorsOptions{})

		require.NoError(t, err)
	})

	t.Run("missing array yields no objects", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.ListMonitors(t.Context(), synthetic.ListMonitorsOptions{})

		require.NoError(t, err)
		assert.Empty(t, resp.Objects)
	})

	t.Run("ignores other top-level fields", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"totalCount":1,"monitors":[{"id":"HTTP_CHECK-1"}]}`))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.ListMonitors(t.Context(), synthetic.ListMonitorsOptions{})

		require.NoError(t, err)
		require.Len(t, resp.Objects, 1)
		assert.JSONEq(t, `{"id":"HTTP_CHECK-1"}`, string(resp.Objects[0]))
	})

	t.Run("non-array monitors field is an error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"monitors":"oops"}`))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.ListMonitors(t.Context(), synthetic.ListMonitorsOptions{})

		assert.ErrorAs(t, err, new(api.RuntimeError))
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.ListMonitors(t.Context(), synthetic.ListMonitorsOptions{})

		assert.ErrorAs(t, err, new(api.APIError))
		assert.Empty(t, resp)
	})

	t.Run("invalid JSON returns RuntimeError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not json"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.ListMonitors(t.Context(), synthetic.ListMonitorsOptions{})

		var rtErr api.RuntimeError
		require.ErrorAs(t, err, &rtErr)
		assert.Equal(t, "synthetic monitors", rtErr.Resource)
		assert.Empty(t, resp)
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.ListMonitors(t.Context(), synthetic.ListMonitorsOptions{})

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic monitors", clientErr.Resource)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Empty(t, resp)
	})
}

func TestGetMonitor(t *testing.T) {
	const response = `{"id":"HTTP_CHECK-1","name":"a"}`

	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors/HTTP_CHECK-1", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.GetMonitor(t.Context(), "HTTP_CHECK-1")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, response, string(resp.Data))
	})

	t.Run("empty id returns error and sends no request", func(t *testing.T) {
		client := synthetic.NewClient(&rest.Client{})

		resp, err := client.GetMonitor(t.Context(), "")

		assert.ErrorIs(t, err, api.ValidationError{Resource: "synthetic monitors", Field: "id", Reason: "is empty"})
		assert.Empty(t, resp)
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors/id", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.GetMonitor(t.Context(), "id")

		assert.Empty(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "boom", string(apiErr.Body))
	})

	t.Run("404 is a not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors/id", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.GetMonitor(t.Context(), "id")

		assert.True(t, api.IsNotFoundError(err))
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.GetMonitor(t.Context(), "id")

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic monitors", clientErr.Resource)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Empty(t, resp)
	})
}

func TestCreateMonitor(t *testing.T) {
	const body = `{"type":"HTTP","name":"a"}`
	const response = `{"monitorId":"HTTP_CHECK-1"}`

	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors", r.URL.Path)
			b, _ := io.ReadAll(r.Body)
			require.Equal(t, body, string(b))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.CreateMonitor(t.Context(), []byte(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, response, string(resp.Data))
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.CreateMonitor(t.Context(), []byte(body))

		assert.Empty(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "boom", string(apiErr.Body))
	})

	t.Run("404 is a not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.CreateMonitor(t.Context(), []byte(body))

		assert.True(t, api.IsNotFoundError(err))
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.CreateMonitor(t.Context(), []byte(body))

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic monitors", clientErr.Resource)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Empty(t, resp)
	})
}

func TestUpdateMonitor(t *testing.T) {
	const body = `{"type":"HTTP","name":"b"}`

	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors/HTTP_CHECK-1", r.URL.Path)
			b, _ := io.ReadAll(r.Body)
			require.Equal(t, body, string(b))
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.UpdateMonitor(t.Context(), "HTTP_CHECK-1", []byte(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("empty id returns error and sends no request", func(t *testing.T) {
		client := synthetic.NewClient(&rest.Client{})

		resp, err := client.UpdateMonitor(t.Context(), "", []byte(body))

		assert.ErrorIs(t, err, api.ValidationError{Resource: "synthetic monitors", Field: "id", Reason: "is empty"})
		assert.Empty(t, resp)
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors/id", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.UpdateMonitor(t.Context(), "id", []byte(body))

		assert.Empty(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "boom", string(apiErr.Body))
	})

	t.Run("404 is a not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors/id", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.UpdateMonitor(t.Context(), "id", []byte(body))

		assert.True(t, api.IsNotFoundError(err))
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.UpdateMonitor(t.Context(), "id", []byte(body))

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic monitors", clientErr.Resource)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Empty(t, resp)
	})
}

func TestDeleteMonitor(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodDelete, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors/HTTP_CHECK-1", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.DeleteMonitor(t.Context(), "HTTP_CHECK-1")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("empty id returns error and sends no request", func(t *testing.T) {
		client := synthetic.NewClient(&rest.Client{})

		resp, err := client.DeleteMonitor(t.Context(), "")

		assert.ErrorIs(t, err, api.ValidationError{Resource: "synthetic monitors", Field: "id", Reason: "is empty"})
		assert.Empty(t, resp)
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodDelete, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors/id", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.DeleteMonitor(t.Context(), "id")

		assert.Empty(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "boom", string(apiErr.Body))
	})

	t.Run("404 is a not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodDelete, r.Method)
			require.Equal(t, "/platform/synthetic/v1/monitors/id", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.DeleteMonitor(t.Context(), "id")

		assert.True(t, api.IsNotFoundError(err))
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.DeleteMonitor(t.Context(), "id")

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic monitors", clientErr.Resource)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Empty(t, resp)
	})
}
