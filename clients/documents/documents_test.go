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

package documents_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
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

func TestDocumentClient_Get(t *testing.T) {
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

	t.Run("Get - no ID given", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "")
		assert.Zero(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("GET - OK", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
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

		multipartTempFileBeforeCount := getTemporaryMultipartFileCount(t)
		resp, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")
		multipartTempFileAfterCount := getTemporaryMultipartFileCount(t)

		assert.NotZero(t, resp)
		assert.Equal(t, "b17ec54b-07ac-4c73-9c4d-232e1b2e2420", resp.ID)
		assert.Equal(t, "my-test-db", resp.Name)
		assert.Equal(t, true, resp.IsPrivate)
		assert.Equal(t, "dashboard", resp.Type)
		assert.Equal(t, 1, resp.Version)
		assert.Equal(t, "12341234-1234-1234-1234-12341234", resp.Owner)
		assert.NotNil(t, resp.OriginAppID)
		assert.Equal(t, "mytest.app", *resp.OriginAppID)
		assert.NotNil(t, resp.OriginExtensionID)
		assert.Equal(t, "mytest.extension", *resp.OriginExtensionID)
		assert.Equal(t, "This is the document content", string(resp.Data))
		assert.NotZero(t, resp.Request)
		assert.Nil(t, err)
		assert.Equal(t, multipartTempFileBeforeCount, multipartTempFileAfterCount, "temporary multipart files should be cleaned up")

	})

	t.Run("GET - no multipart given", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
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

		_, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")
		assert.ErrorIs(t, err, http.ErrNotMultipart)
	})

	t.Run("GET - no boundary given", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
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

		_, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")
		assert.ErrorContains(t, err, "unable to read multipart")
	})

	t.Run("GET - no metadata given", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
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

		multipartTempFileBeforeCount := getTemporaryMultipartFileCount(t)
		_, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")
		multipartTempFileAfterCount := getTemporaryMultipartFileCount(t)

		assert.ErrorIs(t, err, documents.ErrNoMetadata)
		assert.Equal(t, multipartTempFileBeforeCount, multipartTempFileAfterCount, "temporary multipart files should be cleaned up")
	})

	t.Run("GET - no content given", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
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

		multipartTempFileBeforeCount := getTemporaryMultipartFileCount(t)
		_, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")
		multipartTempFileAfterCount := getTemporaryMultipartFileCount(t)

		assert.ErrorIs(t, err, documents.ErrNoContent)
		assert.Equal(t, multipartTempFileBeforeCount, multipartTempFileAfterCount, "temporary multipart files should be cleaned up")
	})

	t.Run("GET - Unable to make HTTP call", func(t *testing.T) {

		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")
		assert.Zero(t, resp)
		assert.Error(t, err)
	})

	t.Run("GET - API Call returned with != 2xx", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusBadRequest,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")
		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusBadRequest, apiError.StatusCode)
	})
}

// getTemporaryMultipartFileCount counts the number of temporary files with the prefix "multipart-".
// These are typically created by multipart reader in the system's temp directory.
// This function is used to ensure that such files are properly cleaned up after processing multipart responses.
func getTemporaryMultipartFileCount(t *testing.T) int {
	entries, err := os.ReadDir(os.TempDir())
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}
	count := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "multipart-") {
			count++
		}
	}
	return count
}

func TestDocumentClient_Create(t *testing.T) {
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

	t.Run("simple case", func(t *testing.T) {

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
						ResponseCode: http.StatusCreated,
						ResponseBody: respPatch,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		res, err := client.Create(t.Context(), documents.Metadata{Name: "name", ID: "extID", Type: documents.Notebook}, []byte("this is the content"))

		require.NoError(t, err)
		assert.JSONEq(t, expected, string(res.Data))
	})

	t.Run("create call returns invalid response body", func(t *testing.T) {

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

		_, err := client.Create(t.Context(), documents.Metadata{Name: "name", ID: "extID", Type: documents.Notebook}, []byte("this is the content"))

		jsonErr := &json.SyntaxError{}
		assert.ErrorAs(t, err, &jsonErr)
	})

	t.Run("create call returns non successful response", func(t *testing.T) {

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

		res, err := client.Create(t.Context(), documents.Metadata{Name: "name", ID: "extID", Type: documents.Notebook}, []byte("this is the content"))

		require.Empty(t, res)
		require.Error(t, err)
	})

	t.Run("patch call returns non successful response; rollback success", func(t *testing.T) {

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

		res, err := client.Create(t.Context(), documents.Metadata{Name: "name", ID: "extID", Type: documents.Notebook}, []byte("this is the content"))

		assert.Empty(t, res)
		assert.Error(t, err)

		// var apiErr api.APIError
		assert.ErrorAs(t, err, &api.APIError{})
		assert.ErrorContains(t, err, "some internal error")
	})

	t.Run("patch call returns 404 - retry succeeds", func(t *testing.T) {

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

		res, err := client.Create(t.Context(), documents.Metadata{Name: "name", ID: "extID", Type: documents.Notebook}, []byte("this is the content"))

		require.NoError(t, err)
		assert.JSONEq(t, expected, string(res.Data))
	})

	t.Run("patch call returns 404 - retry fails", func(t *testing.T) {

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
		for range 5 {
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

		res, err := client.Create(t.Context(), documents.Metadata{Name: "name", ID: "extID", Type: documents.Notebook}, []byte("this is the content"))

		assert.Empty(t, res)
		assert.Error(t, err)

		// var apiErr api.APIError
		assert.ErrorAs(t, err, &api.APIError{})
		assert.ErrorContains(t, err, "failed with status code 404")
	})

	t.Run("patch call returns non successful response; rollback fails", func(t *testing.T) {

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

		res, err := client.Create(t.Context(), documents.Metadata{Name: "name", ID: "extID", Type: documents.Notebook}, []byte("this is the content"))

		assert.Empty(t, res)
		assert.Error(t, err)

		// var apiErr api.APIError
		assert.ErrorAs(t, err, &api.APIError{})
		assert.ErrorContains(t, err, "some internal error")
	})
}

func TestDocumentClient_Update(t *testing.T) {
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

	t.Run("Update - Missing id", func(t *testing.T) {
		responses := []testutils.ResponseDef{}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), documents.Metadata{ID: "", Name: "my-dashboard", IsPrivate: true, Type: documents.Dashboard}, []byte(documentContent))
		assert.Zero(t, resp)
		assert.Error(t, err)
	})

	t.Run("Update - Document not found", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), documents.Metadata{ID: "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", Name: "my-dashboard", IsPrivate: true, Type: documents.Dashboard}, []byte(documentContent))
		assert.Zero(t, resp)
		assert.Error(t, err)
	})

	t.Run("Update - Fails to fetch existing document", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), documents.Metadata{ID: "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", Name: "my-dashboard", IsPrivate: true, Type: documents.Dashboard}, []byte(documentContent))

		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusInternalServerError, apiError.StatusCode)

	})

	t.Run("Update - Existing document found", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  contentType,
					}
				},
			},
			{
				PATCH: func(t *testing.T, req *http.Request) testutils.Response {
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

		resp, err := client.Update(t.Context(), documents.Metadata{ID: "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", Name: "my-dashboard", IsPrivate: true, Type: documents.Dashboard}, []byte(documentContent))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.JSONEq(t, expected, string(resp.Data))
	})

	t.Run("Update - Existing document found with arbitrary document-kind", func(t *testing.T) {
		const someArbitraryDocumentType = "my-super-awesome-document-kind"

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {

					payload := strings.ReplaceAll(getPayload, documents.Dashboard, someArbitraryDocumentType)

					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payload,
						ContentType:  contentType,
					}
				},
			},
			{
				PATCH: func(t *testing.T, req *http.Request) testutils.Response {
					if err := req.ParseMultipartForm(32 << 10); err != nil { // 32 kb
						t.Fatalf("Unable to parse multipart form data: %s", err)
					}

					assert.Equal(t, []string{someArbitraryDocumentType}, req.MultipartForm.Value["type"])
					assert.Equal(t, []string{"false"}, req.MultipartForm.Value["isPrivate"])
					assert.Equal(t, []string{"my-dashboard"}, req.MultipartForm.Value["name"])

					payload := strings.ReplaceAll(patchPayload, documents.Dashboard, someArbitraryDocumentType)

					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payload,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), documents.Metadata{ID: "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", Name: "my-dashboard", Type: someArbitraryDocumentType}, []byte(documentContent))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		expected := strings.ReplaceAll(expected, documents.Dashboard, someArbitraryDocumentType)

		assert.JSONEq(t, expected, string(resp.Data))
	})

	t.Run("Update - Existing document found - Update fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  contentType,
					}
				},
			},
			{
				PATCH: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), documents.Metadata{ID: "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", Name: "my-dashboard", IsPrivate: true, Type: documents.Dashboard}, []byte(documentContent))

		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusInternalServerError, apiError.StatusCode)
	})

	t.Run("Update - Existing document found - Update fails due to invalid response body", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  contentType,
					}
				},
			},
			{
				PATCH: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		_, err := client.Update(t.Context(), documents.Metadata{ID: "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", Name: "my-dashboard", IsPrivate: true, Type: documents.Dashboard}, []byte(documentContent))

		jsonErr := &json.SyntaxError{}
		assert.ErrorAs(t, err, &jsonErr)
	})
}

func TestDocumentClient_List(t *testing.T) {
	// The list endpoint returns a reduced metadata object. These payloads are what
	// it returns for the add-field values List asks for, so they carry description,
	// labels and originExtensionId — but not isReshareable, which cannot be
	// requested and is the reason List still reads each document individually.
	const listPayloadPage1 = `{
    "documents": [
        {
            "id": "id1",
            "name": "name1",
			"isPrivate": true,
            "type": "dashboard",
            "version": 1,
            "owner": "owner1",
			"originExtensionId": null,
			"description": "description1",
			"labels": ["alpha", "beta"]
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
			"originExtensionId": "extension1",
			"labels": []
        }
    ],
    "nextPageKey": null,
    "totalCount": 2
}`

	// wantAddFields is the add-field set List must request; without it the list
	// payloads above would come back without description, labels or
	// originExtensionId.
	const wantAddFields = "add-field=description&add-field=labels&add-field=originExtensionId"

	// metadataPayload wraps a metadata JSON object in the multipart body that the
	// single-document GET returns.
	metadataPayload := func(metadataJSON string) string {
		part := "Content-Disposition: form-data; name=\"metadata\"\nContent-Type: application/json\n\n" + metadataJSON
		return fmt.Sprintf("--%s\n%s\n--%s--", boundary, part, boundary)
	}

	metadataID1 := metadataPayload(`{
    "modificationInfo": {
        "createdBy": "12341234-1234-1234-1234-12341234",
        "createdTime": "2024-04-10T17:21:06.797Z",
        "lastModifiedBy": "2f321c04-566e-4779-b576-3c033b8cd9e9",
        "lastModifiedTime": "2024-04-10T17:21:06.797Z"
    },
    "access": ["read", "write", "delete"],
    "id": "id1",
    "name": "name1",
	"isPrivate": true,
    "type": "dashboard",
    "version": 1,
    "owner": "owner1",
	"originAppId": "app1",
	"originExtensionId": null,
	"description": "description1",
	"isReshareable": false,
	"labels": ["alpha", "beta"]
}`)

	metadataID2 := metadataPayload(`{
    "modificationInfo": {
        "createdBy": "12341234-1234-1234-1234-12341234",
        "createdTime": "2024-04-10T17:21:06.797Z",
        "lastModifiedBy": "2f321c04-566e-4779-b576-3c033b8cd9e9",
        "lastModifiedTime": "2024-04-10T17:21:06.797Z"
    },
    "access": ["read", "write", "delete"],
    "id": "id2",
    "name": "name2",
	"isPrivate": false,
    "type": "dashboard",
    "version": 1,
    "owner": "owner2",
	"originAppId": null,
	"originExtensionId": "extension1",
	"labels": []
}`)

	// documentGET is the response def for the single-document GET of wantID.
	documentGET := func(wantID, body string) testutils.ResponseDef {
		return testutils.ResponseDef{
			GET: func(t *testing.T, req *http.Request) testutils.Response {
				return testutils.Response{
					ResponseCode: http.StatusOK,
					ResponseBody: body,
					ContentType:  contentType,
				}
			},
			ValidateRequest: func(t *testing.T, request *http.Request) {
				assert.True(t, strings.HasSuffix(request.URL.Path, "/documents/"+wantID),
					"expected a single-document GET for %s, got %s", wantID, request.URL.Path)
			},
		}
	}

	t.Run("List - OK", func(t *testing.T) {
		// One document GET per listed id, interleaved with the paged list requests.
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: listPayloadPage1,
					}
				},
				ValidateRequest: func(t *testing.T, request *http.Request) {
					assert.Equal(t, wantAddFields+"&filter=type+%3D%3D+%27dashboard%27", request.URL.RawQuery)
				},
			},
			documentGET("id1", metadataID1),
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: listPayloadPage2,
					}
				},
				ValidateRequest: func(t *testing.T, request *http.Request) {
					assert.Equal(t, wantAddFields+"&filter=type+%3D%3D+%27dashboard%27&page-key=next", request.URL.RawQuery)
				},
			},
			documentGET("id2", metadataID2),
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context(), "type == 'dashboard'")
		assert.NoError(t, err)
		assert.Len(t, resp.Responses, 2)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		assert.Equal(t, "id1", resp.Responses[0].ID)
		assert.Equal(t, "name1", resp.Responses[0].Name)
		assert.Equal(t, true, resp.Responses[0].IsPrivate)
		assert.Equal(t, "dashboard", resp.Responses[0].Type)
		assert.Equal(t, 1, resp.Responses[0].Version)
		assert.Equal(t, "owner1", resp.Responses[0].Owner)
		assert.NotNil(t, resp.Responses[0].OriginAppID)
		assert.Equal(t, "app1", *resp.Responses[0].OriginAppID)
		assert.Nil(t, resp.Responses[0].OriginExtensionID)
		require.NotNil(t, resp.Responses[0].Description)
		assert.Equal(t, "description1", *resp.Responses[0].Description)
		assert.Equal(t, []string{"alpha", "beta"}, resp.Responses[0].Labels)
		require.NotNil(t, resp.Responses[0].IsReshareable)
		assert.False(t, *resp.Responses[0].IsReshareable)
		assert.Equal(t, []string{"read", "write", "delete"}, resp.Responses[0].Access)
		require.NotNil(t, resp.Responses[0].ModificationInfo)
		assert.Equal(t, "12341234-1234-1234-1234-12341234", resp.Responses[0].ModificationInfo.CreatedBy)
		assert.Equal(t, http.StatusOK, resp.Responses[0].StatusCode)

		assert.Equal(t, "id2", resp.Responses[1].ID)
		assert.Equal(t, "name2", resp.Responses[1].Name)
		assert.Equal(t, false, resp.Responses[1].IsPrivate)
		assert.Equal(t, "dashboard", resp.Responses[1].Type)
		assert.Equal(t, 1, resp.Responses[1].Version)
		assert.Equal(t, "owner2", resp.Responses[1].Owner)
		assert.Nil(t, resp.Responses[1].OriginAppID)
		assert.NotNil(t, resp.Responses[1].OriginExtensionID)
		assert.EqualValues(t, "extension1", *resp.Responses[1].OriginExtensionID)
		assert.Nil(t, resp.Responses[1].Description)
		assert.Nil(t, resp.Responses[1].IsReshareable)
		assert.Empty(t, resp.Responses[1].Labels)
		require.NotNil(t, resp.Responses[1].ModificationInfo)
		assert.Equal(t, http.StatusOK, resp.Responses[1].StatusCode)

	})

	t.Run("List - the per document GET wins over the listed metadata", func(t *testing.T) {
		// Both sources can carry labels: the list payload because List asks for them
		// via add-field, and the per document GET because it always returns them.
		// While the GET workaround is in place it is the authoritative one. When
		// isReshareable becomes an add-field value and the GET is dropped, this
		// expectation flips to the listed labels.
		getMetadataWithOtherLabels := metadataPayload(`{
    "id": "id1",
    "name": "name1",
	"isPrivate": true,
    "type": "dashboard",
    "version": 1,
    "owner": "owner1",
	"labels": ["gamma"]
}`)

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: listPayloadPage1,
					}
				},
				ValidateRequest: func(t *testing.T, request *http.Request) {
					assert.Contains(t, request.URL.RawQuery, "add-field=labels",
						"labels must be requested explicitly, the list endpoint omits them otherwise")
				},
			},
			documentGET("id1", getMetadataWithOtherLabels),
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: listPayloadPage2,
					}
				},
			},
			documentGET("id2", metadataID2),
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context(), "type == 'dashboard'")
		require.NoError(t, err)
		require.Len(t, resp.Responses, 2)

		assert.Equal(t, []string{"gamma"}, resp.Responses[0].Labels)
	})

	t.Run("List - Reading a document's metadata fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: listPayloadPage1,
					}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusInternalServerError}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context(), "")

		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusInternalServerError, apiError.StatusCode)
	})

	t.Run("List - Loading Page Fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: listPayloadPage1,
					}
				},
			},
			documentGET("id1", metadataID1),
			// The request for the second page fails.
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context(), "")

		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusInternalServerError, apiError.StatusCode)
		assert.Len(t, resp.Responses, 0)

	})

	t.Run("List - Request Fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.List(t.Context(), "")
		assert.Zero(t, resp)
		assert.Error(t, err)
	})

	t.Run("List - Requested data is invalid", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: "",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		_, err := client.List(t.Context(), "type == 'dashboard'")
		jsonErr := &json.SyntaxError{}
		assert.ErrorAs(t, err, &jsonErr)
	})
}

func TestDocumentClient_Delete(t *testing.T) {

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

	t.Run("Delete - id missing", func(t *testing.T) {
		responses := []testutils.ResponseDef{}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Delete(t.Context(), "")
		assert.Zero(t, resp)
		assert.Error(t, err)

	})

	t.Run("Delete - OK", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					assert.Equal(t, "/platform/document/v1/documents/id-of-document", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  contentType,
					}
				},
			},
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					assert.Equal(t, "/platform/document/v1/documents/id-of-document", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
					}
				},
			},
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					assert.Equal(t, "/platform/document/v1/trash/documents/id-of-document", req.URL.Path)
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
		assert.NotZero(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, err)
	})

	t.Run("Delete - Fails finding existing document", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
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

		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusNotFound, apiError.StatusCode)
	})

	t.Run("Delete - Failed to execute Request", func(t *testing.T) {
		responses := []testutils.ResponseDef{}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Delete(t.Context(), "id-of-document")
		assert.Zero(t, resp)
		assert.Error(t, err)
	})
}
