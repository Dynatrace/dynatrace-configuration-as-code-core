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

package credentialvault_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/credentialvault"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

func TestNewClient(t *testing.T) {
	actual := credentialvault.NewClient(&rest.Client{})
	require.IsType(t, credentialvault.Client{}, *actual)
}

const nextPageKeyParam = "page-key"
const pageSizeParam = "page-size"
const errorResponse = `{
  "error": {
    "code": 400,
    "message": "Invalid request body"
  }
}`

func TestList(t *testing.T) {
	t.Run("successfully returns all credential vault entries across multiple pages", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "credentials": [
    {"name": "name-1", "id": "my-id-1"}
  ]
}`
		apiResponse2 := `{
  "totalCount": 2,
  "credentials": [
    {"name": "name-2", "id": "my-id-2"}
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/credential-vault/v1/entries", req.URL.Path)
					require.Equal(t, "500", req.URL.Query().Get(pageSizeParam))
					require.Equal(t, "", req.URL.Query().Get(nextPageKeyParam))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse1}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/credential-vault/v1/entries", req.URL.Path)
					require.Equal(t, "key_for_next_page", req.URL.Query().Get(nextPageKeyParam))
					require.Equal(t, "", req.URL.Query().Get(pageSizeParam))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse2}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Len(t, resp, 2, "for each call one listResponse should be present")
		assert.Len(t, resp.All(), 2, "two credential vault objects in total should be downloaded")
	})

	t.Run("errors if can't execute all calls successfully", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "credentials": [
    {"name": "my-name", "id": "my-id"}
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse1}
				},
			},
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusInternalServerError, ResponseBody: "Some error message from the server"}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())

		assert.Empty(t, resp)
		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.List(t.Context())

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.ClientError{})
	})

	t.Run("errors if JSON unmarshalling fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: `invalid json`}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.RuntimeError{})
	})
}

func TestGet(t *testing.T) {
	t.Run("successfully returns a credential vault entry for the requested ID", func(t *testing.T) {
		getResponse := `{"name": "my-name", "id": "my-id"}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/credential-vault/v1/entries/my-id", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: getResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "my-id")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, getResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without ID", func(t *testing.T) {
		client := credentialvault.NewClient(&rest.Client{})

		resp, err := client.Get(t.Context(), "")

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "credential-vault", Field: "ID", Reason: "is empty"})
	})

	t.Run("errors if credential vault entry with ID doesn't exist on server", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "false_ID")

		assert.Empty(t, resp)
		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Get(t.Context(), "my-id")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.ClientError{})
	})
}

func TestCreate(t *testing.T) {
	given := `{"name": "my-name"}`

	t.Run("successfully creates a credential vault entry", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/credential-vault/v1/entries", req.URL.Path)
					requestBody, _ := io.ReadAll(req.Body)
					require.JSONEq(t, given, string(requestBody))
					return testutils.Response{ResponseCode: http.StatusCreated, ResponseBody: `{"id": "new-id"}`}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Create(t.Context(), []byte(given))

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("errors if server returns an error", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusBadRequest, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Create(t.Context(), []byte(given))

		assert.Empty(t, resp)
		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Create(t.Context(), []byte(given))

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.ClientError{})
	})
}

func TestUpdate(t *testing.T) {
	given := `{"name": "updated"}`

	t.Run("successfully updates a credential vault entry", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/credential-vault/v1/entries/my-id", req.URL.Path)
					requestBody, _ := io.ReadAll(req.Body)
					require.JSONEq(t, given, string(requestBody))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: `{}`}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "my-id", []byte(given))

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without ID", func(t *testing.T) {
		client := credentialvault.NewClient(&rest.Client{})

		resp, err := client.Update(t.Context(), "", []byte(given))

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "credential-vault", Field: "ID", Reason: "is empty"})
	})

	t.Run("errors if server returns an error", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "false_ID", []byte(given))

		assert.Empty(t, resp)
		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Update(t.Context(), "my-id", []byte(given))

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.ClientError{})
	})
}

func TestDelete(t *testing.T) {
	t.Run("successfully deletes a credential vault entry", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/credential-vault/v1/entries/my-id", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusNoContent}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.Client()))

		_, err := client.Delete(t.Context(), "my-id")

		assert.NoError(t, err)
	})

	t.Run("errors if called without an ID", func(t *testing.T) {
		client := credentialvault.NewClient(&rest.Client{})

		_, err := client.Delete(t.Context(), "")

		assert.ErrorIs(t, err, api.ValidationError{Resource: "credential-vault", Field: "ID", Reason: "is empty"})
	})

	t.Run("errors if credential vault entry with ID doesn't exist on server", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.Client()))

		_, err := client.Delete(t.Context(), "false_ID")

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := credentialvault.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		_, err := client.Delete(t.Context(), "my-id")

		assert.ErrorAs(t, err, &api.ClientError{})
	})
}
