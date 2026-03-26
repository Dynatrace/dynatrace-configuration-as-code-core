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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/directshares"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

func TestNewClient(t *testing.T) {
	actual := directshares.NewClient(&rest.Client{})
	require.IsType(t, directshares.Client{}, *actual)
}

func TestList(t *testing.T) {
	t.Run("successfully returns all direct shares", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "direct-shares": [
    {
      "id": "direct-share-id-1",
      "documentId": "doc-id-1",
      "access": [
        "read",
        "write"
      ],
      "userCount": 2,
      "groupCount": 2
    }
  ]
}`
		apiResponse2 := `{
  "totalCount": 2,
  "direct-shares": [
    {
      "id": "direct-share-id-2",
      "documentId": "doc-id-2",
      "access": [
        "read"
      ],
      "userCount": 2,
      "groupCount": 2
    }
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/direct-shares", req.URL.Path)
					require.Equal(t, "", req.URL.Query().Get("page-key"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse1}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/direct-shares", req.URL.Path)
					require.Equal(t, "key_for_next_page", req.URL.Query().Get("page-key"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse2}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 2, "for each call one listResponse should be present")
		assert.Len(t, resp.All(), 2, "two direct share objects in total should be downloaded")
	})

	t.Run("errors if can't execute all calls successfully", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "direct-shares": [
    {
      "id": "direct-share-id-1",
      "documentId": "doc-id-1",
      "access": [
        "read",
        "write"
      ],
      "userCount": 2,
      "groupCount": 2
    }
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/direct-shares", req.URL.Path)
					require.Equal(t, "", req.URL.Query().Get("page-key"))
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

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.List(t.Context())

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)
	})

	t.Run("errors if JSON unmarshaling fails", func(t *testing.T) {
		invalidJSONResponse := `invalid json`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/direct-shares", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: invalidJSONResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "direct-shares", runtimeErr.Resource)
		assert.Equal(t, "unmarshalling failed", runtimeErr.Reason)
	})
}

func TestGet(t *testing.T) {

	t.Run("successfully returns direct share for requested ID", func(t *testing.T) {
		getResponse := `{
  "id": "direct-share-id-1",
  "documentId": "doc-id-1",
  "access": [
    "read",
    "write"
  ],
  "userCount": 2,
  "groupCount": 2
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/direct-shares/direct-share-id-1", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: getResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "direct-share-id-1")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, getResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := directshares.NewClient(&rest.Client{})

		actual, err := client.Get(t.Context(), "")

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "direct-shares", Field: "id", Reason: "is empty"})

	})

	t.Run("errors if direct share with ID doesn't exist on server", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 404,
    "message": "Provided ID 'false_ID' is not a direct share."
  }
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "false_ID")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)
		assert.Equal(t, "false_ID", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Get(t.Context(), "some-id")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)
		assert.Equal(t, "some-id", clientErr.Identifier)
	})
}

func TestGetRecipients(t *testing.T) {
	t.Run("successfully returns all recipients for given ID", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "recipients": [
	{
	  "type": "user",
	  "id": "user-id-1"
	}
  ]
}`
		apiResponse2 := `{
  "totalCount": 2,
  "recipients": [
	{
	  "type": "group",
	  "id": "group-id-1"
	}
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/direct-shares/direct-share-id-1/recipients", req.URL.Path)
					require.Equal(t, "", req.URL.Query().Get("page-key"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse1}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/direct-shares/direct-share-id-1/recipients", req.URL.Path)
					require.Equal(t, "key_for_next_page", req.URL.Query().Get("page-key"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse2}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetRecipients(t.Context(), "direct-share-id-1")

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 2, "for each call one listResponse should be present")
		assert.Len(t, resp.All(), 2, "two recipient objects in total should be downloaded")
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := directshares.NewClient(&rest.Client{})

		resp, err := client.GetRecipients(t.Context(), "")

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "direct-shares", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if can't execute all calls successfully", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "recipients": [
	{
	  "type": "user",
	  "id": "user-id-1"
	}
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/direct-shares/direct-share-id-1/recipients", req.URL.Path)
					require.Equal(t, "", req.URL.Query().Get("page-key"))
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

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetRecipients(t.Context(), "direct-share-id-1")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if direct share with ID doesn't exist on server", func(t *testing.T) {
		errorResponse := `{
  "error": {
	"code": 404,
	"message": "Provided ID 'false_ID' is not a direct share."
  }
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetRecipients(t.Context(), "false_ID")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.GetRecipients(t.Context(), "some-id")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)
		assert.Equal(t, "some-id", clientErr.Identifier)
	})

	t.Run("errors if JSON unmarshaling fails", func(t *testing.T) {
		invalidJSONResponse := `invalid json`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/direct-shares/direct-share-id-1/recipients", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: invalidJSONResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetRecipients(t.Context(), "direct-share-id-1")
		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "direct-shares", runtimeErr.Resource)
		assert.Equal(t, "unmarshalling failed", runtimeErr.Reason)
	})
}
func TestAddRecipients(t *testing.T) {
	t.Run("successfully adds recipients for given ID", func(t *testing.T) {
		given := `{"recipients":[{"type":"user","id":"user-id-1"}]}`

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/direct-shares/direct-share-id-1/recipients/add", req.URL.Path)
					requestBody, _ := io.ReadAll(req.Body)
					require.JSONEq(t, given, string(requestBody))
					return testutils.Response{ResponseCode: http.StatusOK}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		err := client.AddRecipients(t.Context(), "direct-share-id-1", []byte(given))

		assert.NoError(t, err)
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := directshares.NewClient(&rest.Client{})

		err := client.AddRecipients(t.Context(), "", []byte(`{}`))

		assert.ErrorIs(t, err, api.ValidationError{Resource: "direct-shares", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if server returns an error", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 404,
    "message": "Provided ID 'false_ID' is not a direct share."
  }
}`

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		err := client.AddRecipients(t.Context(), "false_ID", []byte(`{}`))

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)
		assert.Equal(t, "false_ID", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		err := client.AddRecipients(t.Context(), "some-id", []byte(`{}`))

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)
		assert.Equal(t, "some-id", clientErr.Identifier)
	})
}

func TestRemoveRecipients(t *testing.T) {
	t.Run("successfully removes recipients for given ID", func(t *testing.T) {
		given := `{"ids":["user-id-1"]}`

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/direct-shares/direct-share-id-1/recipients/remove", req.URL.Path)
					requestBody, _ := io.ReadAll(req.Body)
					require.JSONEq(t, given, string(requestBody))
					return testutils.Response{ResponseCode: http.StatusOK}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		err := client.RemoveRecipients(t.Context(), "direct-share-id-1", []byte(given))

		assert.NoError(t, err)
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := directshares.NewClient(&rest.Client{})

		err := client.RemoveRecipients(t.Context(), "", []byte(`{}`))

		assert.ErrorIs(t, err, api.ValidationError{Resource: "direct-shares", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if server returns an error", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 403,
    "message": "Not authorized."
  }
}`

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusForbidden, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		err := client.RemoveRecipients(t.Context(), "false_ID", []byte(`{}`))

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)
		assert.Equal(t, "false_ID", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		err := client.RemoveRecipients(t.Context(), "some-id", []byte(`{}`))

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)
		assert.Equal(t, "some-id", clientErr.Identifier)
	})
}

func TestCreate(t *testing.T) {
	given := `{
  "documentId": "doc-id-1",
  "access": "read",
  "recipients": [
	{
	  "type": "user",
	  "id": "user-id-1"
	}
  ]
}`

	t.Run("successfully creates a direct share", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/direct-shares", req.URL.Path)
					requestBody, _ := io.ReadAll(req.Body)
					require.JSONEq(t, given, string(requestBody))
					return testutils.Response{ResponseCode: http.StatusCreated}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.Create(t.Context(), []byte(given))

		assert.NotEmpty(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, actual.StatusCode)
	})

	t.Run("errors if server returns an error", func(t *testing.T) {
		errorResponse := `{
	  "error": {
		"code": 400,
		"message": "Invalid request body"
	  }
	}`

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusBadRequest, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.Create(t.Context(), []byte(given))

		assert.Empty(t, actual)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		actual, err := client.Create(t.Context(), []byte(given))
		assert.Empty(t, actual)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)
	})
}

func TestDelete(t *testing.T) {
	t.Run("successfully deletes entity with ID from server", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/direct-shares/direct-share-id-1", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusNoContent}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		err := client.Delete(t.Context(), "direct-share-id-1")

		assert.NoError(t, err)
	})

	t.Run("errors when called without ID parameter", func(t *testing.T) {
		client := directshares.NewClient(&rest.Client{})

		err := client.Delete(t.Context(), "")

		assert.ErrorIs(t, err, api.ValidationError{Resource: "direct-shares", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if direct share with ID doesn't exist on server", func(t *testing.T) {
		get404Response := `{
  "error": {
    "code": 404,
    "message": "Provided ID 'false_ID' is not a direct share."
  }
}`

		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: get404Response}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.Client()))

		err := client.Delete(t.Context(), "false_ID")

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)
		assert.Equal(t, "false_ID", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, get404Response, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := directshares.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		err := client.Delete(t.Context(), "some-id")

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "direct-shares", clientErr.Resource)
		assert.Equal(t, "some-id", clientErr.Identifier)
	})
}
