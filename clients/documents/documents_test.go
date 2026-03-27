// @license
// Copyright 2023 Dynatrace LLC
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

package documents_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/documents"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

const boundary = "Aas2UU1KdxSpaAyiNZ4-tnuzbwqnKuNK8vMOGy"

var contentType = fmt.Sprintf("multipart/form-data;boundary=%s", boundary)

func TestNewClient(t *testing.T) {
	actual := documents.NewClient(&rest.Client{})
	require.IsType(t, documents.Client{}, *actual)
}

func TestGet(t *testing.T) {
	const metadataContent = `Content-Disposition: form-data; name="metadata"
Content-Type: application/json

{
    "modificationInfo": {
        "createdBy": "12341234-1234-1234-1234-12341234",
        "createdTime": "2024-04-11T12:31:33.599Z",
        "lastModifiedBy": "2f321c04-566e-4779-b576-3c033b8cd9e9",
        "lastModifiedTime": "2024-04-11T12:31:33.599Z"
    },
    "access": [
        "read",
        "delete",
        "write"
    ],
    "id": "b17ec54b-07ac-4c73-9c4d-232e1b2e2420",
    "name": "my-test-db",
	"isPrivate": true,
    "type": "dashboard",
    "version": 1,
    "owner": "12341234-1234-1234-1234-12341234",
	"originAppId": "mytest.app",
	"originExtensionId": "mytest.extension"
}`
	const payloadContent = `Content-Disposition: form-data; name="content"; filename="my-test-db"
Content-Type: application/json

This is the document content`

	payload := fmt.Sprintf("--%s\n%s\n--%s\n%s\n--%s--", boundary, metadataContent, boundary, payloadContent, boundary)
	payloadWithoutMetadata := fmt.Sprintf("--%s\n%s\n--%s--", boundary, payloadContent, boundary)
	payloadWithoutContent := fmt.Sprintf("--%s\n%s\n--%s--", boundary, metadataContent, boundary)

	t.Run("successfully returns document for requested ID", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/documents/b17ec54b-07ac-4c73-9c4d-232e1b2e2420", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payload,
						ContentType:  contentType,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, "This is the document content", string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := documents.NewClient(&rest.Client{})

		resp, err := client.Get(t.Context(), "")

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "documents", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if document with ID doesn't exist on server", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)
		assert.Equal(t, "b17ec54b-07ac-4c73-9c4d-232e1b2e2420", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)
		assert.Equal(t, "b17ec54b-07ac-4c73-9c4d-232e1b2e2420", clientErr.Identifier)
	})

	t.Run("errors if response is not multipart", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payload,
						ContentType:  "application/json",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "documents", runtimeErr.Resource)
		assert.Equal(t, "b17ec54b-07ac-4c73-9c4d-232e1b2e2420", runtimeErr.Identifier)
		assert.Equal(t, "failed to extract multipart boundary", runtimeErr.Reason)
	})

	t.Run("errors if multipart boundary is missing", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payload,
						ContentType:  "multipart/form-data",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "documents", runtimeErr.Resource)
	})

	t.Run("errors if metadata field not found in response", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payloadWithoutMetadata,
						ContentType:  contentType,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "documents", runtimeErr.Resource)
		assert.Equal(t, "metadata field not found in response", runtimeErr.Reason)
	})

	t.Run("errors if content field not found in response", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payloadWithoutContent,
						ContentType:  contentType,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "documents", runtimeErr.Resource)
		assert.Equal(t, "content field not found in response", runtimeErr.Reason)
	})
}

func TestList(t *testing.T) {
	const listPayloadPage1 = `{
    "documents": [
        {
            "id": "id1",
            "name": "name1",
            "isPrivate": true,
            "type": "dashboard",
            "version": 1,
            "owner": "owner1",
            "originAppId": "app1"
        }
    ],
    "nextPageKey": "next",
    "totalCount": 2
}`

	const listPayloadPage2 = `{
    "documents": [
        {
            "id": "id2",
            "name": "name2",
            "isPrivate": false,
            "type": "dashboard",
            "version": 1,
            "owner": "owner2",
            "originExtensionId": "extension1"
        }
    ],
    "nextPageKey": null,
    "totalCount": 2
}`

	t.Run("successfully returns all documents with pagination", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/documents", req.URL.Path)
					require.Equal(t, "add-field=originExtensionId&filter=type+%3D%3D+%27dashboard%27", req.URL.RawQuery)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: listPayloadPage1}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/documents", req.URL.Path)
					require.Equal(t, "add-field=originExtensionId&filter=type+%3D%3D+%27dashboard%27&page-key=next", req.URL.RawQuery)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: listPayloadPage2}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context(), "type == 'dashboard'")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Len(t, resp.Responses, 2, "two document objects in total should be downloaded")
	})

	t.Run("errors if can't execute all calls successfully", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: listPayloadPage1}
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

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context(), "")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.List(t.Context(), "")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)
	})

	t.Run("errors if JSON unmarshaling fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/documents", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: "invalid json"}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context(), "type == 'dashboard'")

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "documents", runtimeErr.Resource)
		assert.Equal(t, "unmarshalling failed", runtimeErr.Reason)
	})
}

func TestCreate(t *testing.T) {
	const (
		expected = `{
  "id": "f0427cd7-c779-4dc1-9cf6-2730738b4ea0",
  "name": "name",
  "type": "notebook",
  "isPrivate": false,
  "description": null,
  "version": 2
}`
		respCreate = `{
  "id": "f6e26fdd-1451-4655-b6ab-1240a00c1fba",
  "name": "name",
  "type": "notebook",
  "isPrivate": true,
  "description": null,
  "version": 1
}`
		respPatch = `{"documentMetadata":` + expected + `}`
	)

	t.Run("successfully creates a new document", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/documents", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusCreated,
						ResponseBody: respCreate,
					}
				},
			},
			{
				PATCH: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusCreated,
						ResponseBody: respPatch,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		res, err := client.Create(t.Context(), "name", false, "extID", []byte("this is the content"), documents.Notebook)

		require.NoError(t, err)
		assert.JSONEq(t, expected, string(res.Data))
	})

	t.Run("errors if POST returns invalid response body", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Create(t.Context(), "name", false, "extID", []byte("this is the content"), documents.Notebook)

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "documents", runtimeErr.Resource)
		assert.Equal(t, "failed to unmarshal create response", runtimeErr.Reason)
	})

	t.Run("errors if POST returns server error", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Create(t.Context(), "name", false, "extID", []byte("this is the content"), documents.Notebook)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Create(t.Context(), "name", false, "extID", []byte("this is the content"), documents.Notebook)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)
	})

	t.Run("rolls back on PATCH failure and returns error", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusCreated,
						ResponseBody: respCreate,
					}
				},
			},
			{
				PATCH: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
						ResponseBody: "some internal error",
					}
				},
			},
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNoContent,
					}
				},
			},
			// trash
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNoContent,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Create(t.Context(), "name", false, "extID", []byte("this is the content"), documents.Notebook)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.ErrorContains(t, err, "some internal error")
	})

	t.Run("retries on PATCH 404 and succeeds", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusCreated,
						ResponseBody: respCreate,
					}
				},
			},
			{
				PATCH: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
					}
				},
			},
			{
				PATCH: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: respPatch,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		res, err := client.Create(t.Context(), "name", false, "extID", []byte("this is the content"), documents.Notebook)

		require.NoError(t, err)
		assert.JSONEq(t, expected, string(res.Data))
	})

	t.Run("errors after exhausting PATCH 404 retries", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusCreated,
						ResponseBody: respCreate,
					}
				},
			},
		}
		for i := 0; i < 5; i++ {
			responses = append(responses, testutils.ResponseDef{
				PATCH: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
					}
				},
			})
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Create(t.Context(), "name", false, "extID", []byte("this is the content"), documents.Notebook)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	})

	t.Run("errors on PATCH failure when rollback also fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusCreated,
						ResponseBody: respCreate,
					}
				},
			},
			{
				PATCH: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
						ResponseBody: "some internal error",
					}
				},
			},
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Create(t.Context(), "name", false, "extID", []byte("this is the content"), documents.Notebook)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)

		assert.ErrorContains(t, err, "some internal error")
	})
}

func TestUpdate(t *testing.T) {
	const (
		expected = `{
"id": "038ab74f-0a3a-4bf8-9068-85e2d633a1e6",
"name": "my-test-db",
"isPrivate": true,
"type": "dashboard",
"version": 1
}`
		getPayload = `--Aas2UU1KdxSpaAyiNZ4-tnuzbwqnKuNK8vMOGy
Content-Disposition: form-data; name="metadata"
Content-Type: application/json

{
    "modificationInfo": {
        "createdBy": "12341234-1234-1234-1234-12341234",
        "createdTime": "2024-04-11T12:31:33.599Z",
        "lastModifiedBy": "2f321c04-566e-4779-b576-3c033b8cd9e9",
        "lastModifiedTime": "2024-04-11T12:31:33.599Z"
    },
    "access": [
        "read",
        "delete",
        "write"
    ],
    "id": "b17ec54b-07ac-4c73-9c4d-232e1b2e2420",
    "name": "my-test-db",
	"isPrivate": true,
    "type": "dashboard",
    "version": 1,
    "owner": "12341234-1234-1234-1234-12341234"
}
--Aas2UU1KdxSpaAyiNZ4-tnuzbwqnKuNK8vMOGy
Content-Disposition: form-data; name="content"; filename="my-test-db"
Content-Type: application/json

This is the document content
--Aas2UU1KdxSpaAyiNZ4-tnuzbwqnKuNK8vMOGy--
`

		documentContent = "This is the document content"

		patchPayload = `{"documentMetadata":` + expected + `}`
	)

	t.Run("successfully updates an existing document", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/documents/038ab74f-0a3a-4bf8-9068-85e2d633a1e6", request.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  contentType,
					}
				},
			},
			{
				PATCH: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/documents/038ab74f-0a3a-4bf8-9068-85e2d633a1e6", req.URL.Path)
					require.Equal(t, "1", req.URL.Query().Get("optimistic-locking-version"))
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: patchPayload,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", true, []byte(documentContent), documents.Dashboard)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.JSONEq(t, expected, string(resp.Data))
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := documents.NewClient(&rest.Client{})

		resp, err := client.Update(t.Context(), "", "my-dashboard", true, []byte(documentContent), documents.Dashboard)

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "documents", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if document with ID doesn't exist on server", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", true, []byte(documentContent), documents.Dashboard)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)
		assert.Equal(t, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails on GET", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Update(t.Context(), "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", true, []byte(documentContent), documents.Dashboard)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)
		assert.Equal(t, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", clientErr.Identifier)
	})

	t.Run("errors if server returns error on GET", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", true, []byte(documentContent), documents.Dashboard)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)
		assert.Equal(t, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if server returns error on PATCH", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  contentType,
					}
				},
			},
			{
				PATCH: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", true, []byte(documentContent), documents.Dashboard)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPatch, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)
		assert.Equal(t, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if PATCH returns invalid response body", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  contentType,
					}
				},
			},
			{
				PATCH: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		_, err := client.Update(t.Context(), "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", true, []byte(documentContent), documents.Dashboard)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "documents", runtimeErr.Resource)
		assert.Equal(t, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", runtimeErr.Identifier)
		assert.Equal(t, "extracting metadata failed", runtimeErr.Reason)
	})

	t.Run("successfully updates with arbitrary document type", func(t *testing.T) {
		const someArbitraryDocumentType = "my-super-awesome-document-kind"

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  contentType,
					}
				},
			},
			{
				PATCH: func(t *testing.T, req *http.Request) testutils.Response {
					if err := req.ParseMultipartForm(32 << 10); err != nil {
						t.Fatalf("Unable to parse multipart form data: %s", err)
					}
					assert.Equal(t, []string{someArbitraryDocumentType}, req.MultipartForm.Value["type"])
					assert.Equal(t, []string{"false"}, req.MultipartForm.Value["isPrivate"])
					assert.Equal(t, []string{"my-dashboard"}, req.MultipartForm.Value["name"])

					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: `{"documentMetadata":{"id":"038ab74f","name":"my-dashboard","type":"` + someArbitraryDocumentType + `","isPrivate":false,"version":1}}`,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", false, []byte(documentContent), someArbitraryDocumentType)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestDelete(t *testing.T) {
	const getPayload = `--Aas2UU1KdxSpaAyiNZ4-tnuzbwqnKuNK8vMOGy
Content-Disposition: form-data; name="metadata"
Content-Type: application/json

{
    "modificationInfo": {
        "createdBy": "12341234-1234-1234-1234-12341234",
        "createdTime": "2024-04-11T12:31:33.599Z",
        "lastModifiedBy": "2f321c04-566e-4779-b576-3c033b8cd9e9",
        "lastModifiedTime": "2024-04-11T12:31:33.599Z"
    },
    "access": [
        "read",
        "delete",
        "write"
    ],
    "id": "b17ec54b-07ac-4c73-9c4d-232e1b2e2420",
    "name": "my-test-db",
	"isPrivate": true,
    "type": "dashboard",
    "version": 1,
    "owner": "12341234-1234-1234-1234-12341234"
}
--Aas2UU1KdxSpaAyiNZ4-tnuzbwqnKuNK8vMOGy
Content-Disposition: form-data; name="content"; filename="my-test-db"
Content-Type: application/json

This is the document content
--Aas2UU1KdxSpaAyiNZ4-tnuzbwqnKuNK8vMOGy--
`

	t.Run("successfully deletes document with ID from server", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/documents/id-of-document", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  contentType,
					}
				},
			},
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/documents/id-of-document", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
					}
				},
			},
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/document/v1/trash/documents/id-of-document", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Delete(t.Context(), "id-of-document")

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := documents.NewClient(&rest.Client{})

		resp, err := client.Delete(t.Context(), "")

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "documents", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if document with ID doesn't exist on server", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Delete(t.Context(), "id-of-document")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)
		assert.Equal(t, "id-of-document", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Delete(t.Context(), "id-of-document")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "documents", clientErr.Resource)
		assert.Equal(t, "id-of-document", clientErr.Identifier)
	})
}
