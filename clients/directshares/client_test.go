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

package directshares_test

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
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/directshares"
)

func TestNewClient(t *testing.T) {
	actual := directshares.NewClient(&rest.Client{})
	require.IsType(t, directshares.Client{}, *actual)
}

func TestList(t *testing.T) {
	t.Run("successfully returned all configuration from server", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "directShares": [
    {
      "id": "direct-share-id-1",
      "documentId": "doc-id-1",
      "accessRole": "viewer"
    }
  ]
}`
		apiResponse2 := `{
  "totalCount": 2,
  "directShares": [
    {
      "id": "direct-share-id-2",
      "documentId": "doc-id-2",
      "accessRole": "editor"
    }
  ]
}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			require.Equal(t, http.MethodGet, request.Method)
			require.Equal(t, "/platform/document/v1/direct-shares", request.URL.Path)

			switch v := request.URL.Query().Get("page-key"); v {
			case "":
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(apiResponse1))
			case "key_for_next_page":
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(apiResponse2))
			default:
				require.Failf(t, "unexpected call", "unexpected call with page-key= %s", v)
			}
		}))
		defer server.Close()

		u, err := url.Parse(server.URL)
		require.NoError(t, err)
		client := directshares.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.List(t.Context())

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 2, "for each call one listResponse should be present")
		assert.Len(t, resp.All(), 2, "two direct share objects in total should be downloaded")
	})

	t.Run("fails if can't execute all calls successfully", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "directShares": [
    {
      "id": "direct-share-id-1",
      "documentId": "doc-id-1",
      "accessRole": "viewer"
    }
  ]
}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			require.Equal(t, http.MethodGet, request.Method)
			require.Equal(t, "/platform/document/v1/direct-shares", request.URL.Path)

			switch v := request.URL.Query().Get("page-key"); v {
			case "":
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(apiResponse1))
			case "key_for_next_page": // provoke error
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte("Some error message from the server"))
			default:
				require.Failf(t, "unexpected call", "unexpected call with page-key= %s", v)
			}
		}))
		defer server.Close()
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		client := directshares.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.List(t.Context())

		assert.Error(t, err)
		assert.ErrorAs(t, err, new(api.APIError))
		assert.Empty(t, resp)
	})
}

func TestGet(t *testing.T) {
	t.Run("when called without id parameter, returns an error", func(t *testing.T) {
		client := directshares.NewClient(&rest.Client{})

		actual, err := client.Get(t.Context(), "")

		assert.Error(t, err)
		assert.Empty(t, actual)
	})

	t.Run("successful request for requested ID", func(t *testing.T) {
		getResponse := `{
  "id": "direct-share-id-1",
  "documentId": "doc-id-1",
  "accessRole": "viewer"
}`
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			t.Log(request.URL.String())
			require.Equal(t, http.MethodGet, request.Method)
			require.Equal(t, "/platform/document/v1/direct-shares/direct-share-id-1", request.URL.Path)

			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte(getResponse))
		}))
		u, err := url.Parse(server.URL)
		require.NoError(t, err)

		client := directshares.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.Get(t.Context(), "direct-share-id-1")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, getResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("if direct share with ID doesn't exist on server returns error", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 404,
    "message": "Provided ID 'false_ID' is not a direct share."
  }
}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte(errorResponse))
		}))
		defer server.Close()

		u, err := url.Parse(server.URL)
		require.NoError(t, err)
		client := directshares.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.Get(t.Context(), "false_ID")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})
}

func TestGetRecipients(t *testing.T) {
	t.Run("when called without id parameter, returns an error", func(t *testing.T) {
		client := directshares.NewClient(&rest.Client{})

		actual, err := client.GetRecipients(t.Context(), "")

		assert.Error(t, err)
		assert.Empty(t, actual)
	})

	t.Run("successful request for recipients of given ID", func(t *testing.T) {
		getResponse := `{
  "recipients": [
    {
      "type": "user",
      "id": "user-id-1"
    }
  ]
}`
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			t.Log(request.URL.String())
			require.Equal(t, http.MethodGet, request.Method)
			require.Equal(t, "/platform/document/v1/direct-shares/direct-share-id-1/recipients", request.URL.Path)

			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte(getResponse))
		}))
		defer server.Close()

		u, err := url.Parse(server.URL)
		require.NoError(t, err)
		client := directshares.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.GetRecipients(t.Context(), "direct-share-id-1")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, getResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("if direct share with ID doesn't exist on server returns error", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 404,
    "message": "Provided ID 'false_ID' is not a direct share."
  }
}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte(errorResponse))
		}))
		defer server.Close()

		u, err := url.Parse(server.URL)
		require.NoError(t, err)
		client := directshares.NewClient(rest.NewClient(u, server.Client()))

		resp, err := client.GetRecipients(t.Context(), "false_ID")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})
}

func TestAddRecipients(t *testing.T) {
	t.Run("when called without id parameter, returns an error", func(t *testing.T) {
		client := directshares.NewClient(&rest.Client{})

		err := client.AddRecipients(t.Context(), "", []byte(`{}`))

		assert.Error(t, err)
	})

	t.Run("successfully adds recipients for given ID", func(t *testing.T) {
		given := `{"recipients":[{"type":"user","id":"user-id-1"}]}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			t.Log(request.URL.String())
			require.Equal(t, http.MethodPost, request.Method)
			require.Equal(t, "/platform/document/v1/direct-shares/direct-share-id-1/recipients/add", request.URL.Path)
			requestBody, _ := io.ReadAll(request.Body)
			require.JSONEq(t, given, string(requestBody))

			writer.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := directshares.NewClient(rest.NewClient(u, server.Client()))

		err := client.AddRecipients(t.Context(), "direct-share-id-1", []byte(given))

		assert.NoError(t, err)
	})

	t.Run("if server returns an error, returns error", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 404,
    "message": "Provided ID 'false_ID' is not a direct share."
  }
}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte(errorResponse))
		}))
		defer server.Close()

		u, err := url.Parse(server.URL)
		require.NoError(t, err)
		client := directshares.NewClient(rest.NewClient(u, server.Client()))

		err = client.AddRecipients(t.Context(), "false_ID", []byte(`{}`))
		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})
}

func TestRemoveRecipients(t *testing.T) {
	t.Run("when called without id parameter, returns an error", func(t *testing.T) {
		client := directshares.NewClient(&rest.Client{})

		err := client.RemoveRecipients(t.Context(), "", []byte(`{}`))

		assert.Error(t, err)
	})

	t.Run("successfully removes recipients for given ID", func(t *testing.T) {
		given := `{"recipients":[{"type":"user","id":"user-id-1"}]}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			t.Log(request.URL.String())
			require.Equal(t, http.MethodPost, request.Method)
			require.Equal(t, "/platform/document/v1/direct-shares/direct-share-id-1/recipients/remove", request.URL.Path)
			requestBody, _ := io.ReadAll(request.Body)
			require.JSONEq(t, given, string(requestBody))

			writer.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		u, err := url.Parse(server.URL)
		require.NoError(t, err)
		client := directshares.NewClient(rest.NewClient(u, server.Client()))

		err = client.RemoveRecipients(t.Context(), "direct-share-id-1", []byte(given))

		assert.NoError(t, err)
	})

	t.Run("if server returns an error, returns error", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 404,
    "message": "Provided ID 'false_ID' is not a direct share."
  }
}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte(errorResponse))
		}))
		defer server.Close()

		u, err := url.Parse(server.URL)
		require.NoError(t, err)
		client := directshares.NewClient(rest.NewClient(u, server.Client()))

		err = client.RemoveRecipients(t.Context(), "false_ID", []byte(`{}`))

		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})
}

func TestCreate(t *testing.T) {
	given := `{
  "documentId": "doc-id-1",
  "accessRole": "viewer"
}`

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		t.Log(request.URL.String())
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, "/platform/document/v1/direct-shares", request.URL.Path)
		requestBody, _ := io.ReadAll(request.Body)
		require.JSONEq(t, given, string(requestBody))

		writer.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	client := directshares.NewClient(rest.NewClient(u, server.Client()))

	actual, err := client.Create(t.Context(), []byte(given))

	assert.NoError(t, err)
	assert.NotEmpty(t, actual)
	assert.Equal(t, http.StatusCreated, actual.StatusCode)
}

func TestDelete(t *testing.T) {
	t.Run("when called without id parameter, returns an error", func(t *testing.T) {
		client := directshares.NewClient(&rest.Client{})

		err := client.Delete(t.Context(), "")

		assert.Error(t, err)
	})

	t.Run("successfully deleted entity with ID from server", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			require.Equal(t, http.MethodDelete, request.Method)
			require.Equal(t, "/platform/document/v1/direct-shares/direct-share-id-1", request.URL.Path)

			writer.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := directshares.NewClient(rest.NewClient(u, server.Client()))

		err := client.Delete(t.Context(), "direct-share-id-1")

		assert.NoError(t, err)
	})

	t.Run("if direct share with ID doesn't exist on server returns an error", func(t *testing.T) {
		get404Response := `{
  "error": {
    "code": 404,
    "message": "Provided ID 'false_ID' is not a direct share."
  }
}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte(get404Response))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := directshares.NewClient(rest.NewClient(u, server.Client()))

		err := client.Delete(t.Context(), "false_ID")

		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, get404Response, string(apiErr.Body))
	})
}
