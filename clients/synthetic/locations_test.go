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

func TestListLocations(t *testing.T) {
	const response = `{"locations":[{"entityId":"SYNTHETIC_LOCATION-1","name":"a"},{"entityId":"SYNTHETIC_LOCATION-2","name":"b"}]}`

	t.Run("extracts locations and sends filters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations", r.URL.Path)
			require.Equal(t, "AWS", r.URL.Query().Get("cloud-platform"))
			require.Equal(t, "PRIVATE", r.URL.Query().Get("type"))
			require.Equal(t, "BROWSER", r.URL.Query().Get("capability"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.ListLocations(t.Context(), synthetic.ListLocationsOptions{CloudPlatform: "AWS", Type: "PRIVATE", Capability: "BROWSER"})

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, resp.Objects, 2)
		assert.JSONEq(t, `{"entityId":"SYNTHETIC_LOCATION-1","name":"a"}`, string(resp.Objects[0]))
	})

	t.Run("omits filters when empty", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations", r.URL.Path)
			require.Empty(t, r.URL.RawQuery)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.ListLocations(t.Context(), synthetic.ListLocationsOptions{})

		require.NoError(t, err)
	})

	t.Run("sends only set filters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations", r.URL.Path)
			require.Equal(t, "PUBLIC", r.URL.Query().Get("type"))
			require.NotContains(t, r.URL.Query(), "cloud-platform")
			require.NotContains(t, r.URL.Query(), "capability")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.ListLocations(t.Context(), synthetic.ListLocationsOptions{Type: "PUBLIC"})

		require.NoError(t, err)
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.ListLocations(t.Context(), synthetic.ListLocationsOptions{})

		assert.ErrorAs(t, err, new(api.APIError))
		assert.Empty(t, resp)
	})

	t.Run("invalid JSON returns RuntimeError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not json"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.ListLocations(t.Context(), synthetic.ListLocationsOptions{})

		var rtErr api.RuntimeError
		require.ErrorAs(t, err, &rtErr)
		assert.Equal(t, "synthetic locations", rtErr.Resource)
		assert.Empty(t, resp)
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.ListLocations(t.Context(), synthetic.ListLocationsOptions{})

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic locations", clientErr.Resource)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Empty(t, resp)
	})
}

func TestGetLocation(t *testing.T) {
	const response = `{"entityId":"SYNTHETIC_LOCATION-1"}`

	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/SYNTHETIC_LOCATION-1", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.GetLocation(t.Context(), "SYNTHETIC_LOCATION-1")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, response, string(resp.Data))
	})

	t.Run("empty id returns error and sends no request", func(t *testing.T) {
		client := synthetic.NewClient(&rest.Client{})

		resp, err := client.GetLocation(t.Context(), "")

		assert.ErrorIs(t, err, api.ValidationError{Resource: "synthetic locations", Field: "id", Reason: "is empty"})
		assert.Empty(t, resp)
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/id", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.GetLocation(t.Context(), "id")

		assert.Empty(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "boom", string(apiErr.Body))
	})

	t.Run("404 is a not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/id", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.GetLocation(t.Context(), "id")

		assert.True(t, api.IsNotFoundError(err))
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.GetLocation(t.Context(), "id")

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic locations", clientErr.Resource)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Empty(t, resp)
	})
}

func TestCreateLocation(t *testing.T) {
	const body = `{"name":"loc","type":"PRIVATE"}`
	const response = `{"entityId":"SYNTHETIC_LOCATION-1"}`

	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations", r.URL.Path)
			b, _ := io.ReadAll(r.Body)
			require.Equal(t, body, string(b))
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.CreateLocation(t.Context(), []byte(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, response, string(resp.Data))
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.CreateLocation(t.Context(), []byte(body))

		assert.Empty(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "boom", string(apiErr.Body))
	})

	t.Run("404 is a not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.CreateLocation(t.Context(), []byte(body))

		assert.True(t, api.IsNotFoundError(err))
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.CreateLocation(t.Context(), []byte(body))

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic locations", clientErr.Resource)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Empty(t, resp)
	})
}

func TestUpdateLocation(t *testing.T) {
	const body = `{"name":"loc2"}`

	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/SYNTHETIC_LOCATION-1", r.URL.Path)
			b, _ := io.ReadAll(r.Body)
			require.Equal(t, body, string(b))
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.UpdateLocation(t.Context(), "SYNTHETIC_LOCATION-1", []byte(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("empty id returns error and sends no request", func(t *testing.T) {
		client := synthetic.NewClient(&rest.Client{})

		resp, err := client.UpdateLocation(t.Context(), "", []byte(body))

		assert.ErrorIs(t, err, api.ValidationError{Resource: "synthetic locations", Field: "id", Reason: "is empty"})
		assert.Empty(t, resp)
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/id", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.UpdateLocation(t.Context(), "id", []byte(body))

		assert.Empty(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "boom", string(apiErr.Body))
	})

	t.Run("404 is a not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/id", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.UpdateLocation(t.Context(), "id", []byte(body))

		assert.True(t, api.IsNotFoundError(err))
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.UpdateLocation(t.Context(), "id", []byte(body))

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic locations", clientErr.Resource)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Empty(t, resp)
	})
}

func TestDeleteLocation(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodDelete, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/SYNTHETIC_LOCATION-1", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.DeleteLocation(t.Context(), "SYNTHETIC_LOCATION-1")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("empty id returns error and sends no request", func(t *testing.T) {
		client := synthetic.NewClient(&rest.Client{})

		resp, err := client.DeleteLocation(t.Context(), "")

		assert.ErrorIs(t, err, api.ValidationError{Resource: "synthetic locations", Field: "id", Reason: "is empty"})
		assert.Empty(t, resp)
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodDelete, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/id", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.DeleteLocation(t.Context(), "id")

		assert.Empty(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "boom", string(apiErr.Body))
	})

	t.Run("404 is a not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodDelete, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/id", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.DeleteLocation(t.Context(), "id")

		assert.True(t, api.IsNotFoundError(err))
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.DeleteLocation(t.Context(), "id")

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic locations", clientErr.Resource)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Empty(t, resp)
	})
}

func TestGetPublicLocationsStatus(t *testing.T) {
	const response = `{"enabled":true}`

	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/status", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.GetPublicLocationsStatus(t.Context())
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, response, string(resp.Data))
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/status", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.GetPublicLocationsStatus(t.Context())

		assert.Empty(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "boom", string(apiErr.Body))
	})

	t.Run("404 is a not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/status", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.GetPublicLocationsStatus(t.Context())

		assert.True(t, api.IsNotFoundError(err))
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.GetPublicLocationsStatus(t.Context())

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic public locations status", clientErr.Resource)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Empty(t, resp)
	})
}

func TestUpdatePublicLocationsStatus(t *testing.T) {
	const body = `{"enabled":false}`

	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/status", r.URL.Path)
			b, _ := io.ReadAll(r.Body)
			require.Equal(t, body, string(b))
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.UpdatePublicLocationsStatus(t.Context(), []byte(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("non-2xx returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/status", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.UpdatePublicLocationsStatus(t.Context(), []byte(body))

		assert.Empty(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "boom", string(apiErr.Body))
	})

	t.Run("404 is a not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/status", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{}"))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.UpdatePublicLocationsStatus(t.Context(), []byte(body))

		assert.True(t, api.IsNotFoundError(err))
	})

	t.Run("transport error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))
		server.Close()

		resp, err := client.UpdatePublicLocationsStatus(t.Context(), []byte(body))

		var clientErr api.ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, "synthetic public locations status", clientErr.Resource)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Empty(t, resp)
	})

	t.Run("409 conflict returns APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/platform/synthetic/v1/locations/status", r.URL.Path)
			w.WriteHeader(http.StatusConflict)
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := synthetic.NewClient(rest.NewClient(u, server.Client()))

		_, err := client.UpdatePublicLocationsStatus(t.Context(), []byte(body))

		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusConflict, apiErr.StatusCode)
	})
}
