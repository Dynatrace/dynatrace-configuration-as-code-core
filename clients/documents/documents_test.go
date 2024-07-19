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
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	docApi "github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/documents"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/documents"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDocumentClient_Get(t *testing.T) {
	const payload = `--Aas2UU1KdxSpaAyiNZ4-tnuzbwqnKuNK8vMOGy
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
	"externalId": "extId",
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

	t.Run("Get - no ID given", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, "")
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
						ContentType:  "multipart/form-data;boundary=Aas2UU1KdxSpaAyiNZ4-tnuzbwqnKuNK8vMOGy",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")
		assert.NotZero(t, resp)
		assert.Equal(t, "b17ec54b-07ac-4c73-9c4d-232e1b2e2420", resp.ID)
		assert.Equal(t, "my-test-db", resp.Name)
		assert.Equal(t, true, resp.IsPrivate)
		assert.Equal(t, "extId", resp.ExternalID)
		assert.Equal(t, "dashboard", resp.Type)
		assert.Equal(t, 1, resp.Version)
		assert.Equal(t, "12341234-1234-1234-1234-12341234", resp.Owner)
		assert.Equal(t, "This is the document content", string(resp.Data))
		assert.NotZero(t, resp.Request)
		assert.Nil(t, err)

	})

	t.Run("GET - Unable to make HTTP call", func(t *testing.T) {

		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")
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
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, "b17ec54b-07ac-4c73-9c4d-232e1b2e2420")
		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusBadRequest, apiError.StatusCode)
	})
}

func TestDocumentClient_Create(t *testing.T) {
	givenDoc := docApi.Document{
		Kind:       "notebook",
		Name:       "name",
		ExternalID: "extID",
		Public:     true,
		Content:    []byte("this is the content"),
	}
	const (
		expected = `{
  "id": "f0427cd7-c779-4dc1-9cf6-2730738b4ea0",
  "name": "name",
  "type": "notebook",
  "isPrivate": false,
  "externalId": "externalID",
  "description": null,
  "version": 2
}`
		respCreate = `{
  "id": "f6e26fdd-1451-4655-b6ab-1240a00c1fba",
  "name": "name",
  "type": "notebook",
  "isPrivate": true,
  "externalId": "externalID",
  "description": null,
  "version": 1
}`
		respPatch = `{"documentMetadata":` + expected + `}`
	)

	t.Run("simple case", func(t *testing.T) {
		ctx := testutils.ContextWithLogger(t)
		mockClient := documents.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().Create(ctx, givenDoc).
			Return(&http.Response{Status: http.StatusText(http.StatusCreated), StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(respCreate)), Request: &http.Request{Method: http.MethodGet, URL: &url.URL{}}}, nil)
		mockClient.EXPECT().Patch(ctx, "f6e26fdd-1451-4655-b6ab-1240a00c1fba", 1, givenDoc).
			Return(&http.Response{Status: http.StatusText(http.StatusOK), StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(respPatch)), Request: &http.Request{Method: http.MethodPatch, URL: &url.URL{}}}, nil)

		docClient := documents.NewTestClient(mockClient)

		res, err := docClient.Create(ctx, "name", false, "extID", []byte("this is the content"), documents.Notebook)

		require.NoError(t, err)
		assert.JSONEq(t, expected, string(res.Data))
	})

	t.Run("create call returns non successful response", func(t *testing.T) {
		ctx := testutils.ContextWithLogger(t)
		mockClient := documents.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().Create(ctx, givenDoc).
			Return(&http.Response{Status: http.StatusText(http.StatusInternalServerError), StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader("some internal error")), Request: &http.Request{Method: http.MethodPost, URL: &url.URL{}}}, nil)

		docClient := documents.NewTestClient(mockClient)

		res, err := docClient.Create(ctx, "name", false, "extID", []byte("this is the content"), documents.Notebook)

		require.Empty(t, res)
		require.Error(t, err)
	})

	t.Run("patch call returns non successful response; rollback success", func(t *testing.T) {
		ctx := testutils.ContextWithLogger(t)
		mockClient := documents.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().Create(ctx, givenDoc).
			Return(&http.Response{Status: http.StatusText(http.StatusCreated), StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(respCreate)), Request: &http.Request{Method: http.MethodGet, URL: &url.URL{}}}, nil)
		mockClient.EXPECT().Patch(ctx, "f6e26fdd-1451-4655-b6ab-1240a00c1fba", 1, givenDoc).
			Return(&http.Response{Status: http.StatusText(http.StatusInternalServerError), StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader("some internal error")), Request: &http.Request{Method: http.MethodPatch, URL: &url.URL{}}}, nil)
		mockClient.EXPECT().Delete(ctx, "f6e26fdd-1451-4655-b6ab-1240a00c1fba", 1).
			Return(&http.Response{Status: http.StatusText(http.StatusNoContent), StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader("")), Request: &http.Request{Method: http.MethodDelete, URL: &url.URL{}}}, nil)
		mockClient.EXPECT().Trash(ctx, "f6e26fdd-1451-4655-b6ab-1240a00c1fba").
			Return(&http.Response{Status: http.StatusText(http.StatusNoContent), StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader("")), Request: &http.Request{Method: http.MethodDelete, URL: &url.URL{}}}, nil)

		docClient := documents.NewTestClient(mockClient)

		res, err := docClient.Create(ctx, "name", false, "extID", []byte("this is the content"), documents.Notebook)

		require.Empty(t, res)
		require.Error(t, err)

		// var apiErr api.APIError
		assert.ErrorAs(t, err, &api.APIError{})
		assert.ErrorContains(t, err, "some internal error")
	})

	t.Run("patch call returns non successful response; rollback fails", func(t *testing.T) {
		ctx := testutils.ContextWithLogger(t)
		mockClient := documents.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().Create(ctx, givenDoc).
			Return(&http.Response{Status: http.StatusText(http.StatusCreated), StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(respCreate)), Request: &http.Request{Method: http.MethodGet, URL: &url.URL{}}}, nil)
		mockClient.EXPECT().Patch(ctx, "f6e26fdd-1451-4655-b6ab-1240a00c1fba", 1, givenDoc).
			Return(&http.Response{Status: http.StatusText(http.StatusInternalServerError), StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader("some internal error")), Request: &http.Request{Method: http.MethodPatch, URL: &url.URL{}}}, nil)
		mockClient.EXPECT().Delete(ctx, "f6e26fdd-1451-4655-b6ab-1240a00c1fba", 1).
			Return(&http.Response{Status: http.StatusText(http.StatusInternalServerError), StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader("")), Request: &http.Request{Method: http.MethodDelete, URL: &url.URL{}}}, nil)

		docClient := documents.NewTestClient(mockClient)

		res, err := docClient.Create(ctx, "name", false, "extID", []byte("this is the content"), documents.Notebook)

		require.Empty(t, res)
		require.Error(t, err)

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
"externalId": "extId",
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
	"externalId": "extId",
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

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "", "my-dashboard", true, []byte(documentContent), documents.Dashboard)
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

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", true, []byte(documentContent), documents.Dashboard)
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

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", true, []byte(documentContent), documents.Dashboard)

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
						ContentType:  "multipart/form-data;boundary=Aas2UU1KdxSpaAyiNZ4-tnuzbwqnKuNK8vMOGy",
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

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", true, []byte(documentContent), documents.Dashboard)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.JSONEq(t, expected, string(resp.Data))
	})

	t.Run("Update - Existing document found - Update fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  "multipart/form-data;boundary=Aas2UU1KdxSpaAyiNZ4-tnuzbwqnKuNK8vMOGy",
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

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", true, []byte(documentContent), documents.Dashboard)

		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusInternalServerError, apiError.StatusCode)
	})
}

func TestDocumentClient_List(t *testing.T) {
	const listPayloadPage1 = `{
    "documents": [
        {
            "modificationInfo": {
                "createdBy": "12341234-1234-1234-1234-12341234",
                "createdTime": "2024-04-10T17:21:06.797Z",
                "lastModifiedBy": "2f321c04-566e-4779-b576-3c033b8cd9e9",
                "lastModifiedTime": "2024-04-10T17:21:06.797Z"
            },
            "access": [
                "read",
                "write",
                "delete"
            ],
            "id": "id1",
            "name": "name1",
			"isPrivate": true,
			"externalId": "extId1",
            "type": "dashboard",
            "version": 1,
            "owner": "owner1"
        }
    ],
    "nextPageKey": "next",
    "totalCount": 2
}`

	const listPayloadPage2 = `{
    "documents": [
        {
            "modificationInfo": {
                "createdBy": "12341234-1234-1234-1234-12341234",
                "createdTime": "2024-04-10T17:21:06.797Z",
                "lastModifiedBy": "2f321c04-566e-4779-b576-3c033b8cd9e9",
                "lastModifiedTime": "2024-04-10T17:21:06.797Z"
            },
            "access": [
                "read",
                "write",
                "delete"
            ],
            "id": "id2",
            "name": "name2",
			"isPrivate": false,
			"externalId": "extId2",
            "type": "dashboard",
            "version": 1,
            "owner": "owner2"
        }
    ],
    "nextPageKey": null,
    "totalCount": 2
}`

	t.Run("List - OK", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: listPayloadPage1,
					}
				},
				ValidateRequest: func(t *testing.T, request *http.Request) {
					assert.Equal(t, "filter=type+%3D%3D+%27dashboard%27", request.URL.RawQuery)
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: listPayloadPage2,
					}
				},
				ValidateRequest: func(t *testing.T, request *http.Request) {
					assert.Equal(t, "filter=type+%3D%3D+%27dashboard%27&page-key=next", request.URL.RawQuery)
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.List(ctx, "type == 'dashboard'")
		assert.Len(t, resp.Responses, 2)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, err)

		assert.Equal(t, "id1", resp.Responses[0].ID)
		assert.Equal(t, "name1", resp.Responses[0].Name)
		assert.Equal(t, true, resp.Responses[0].IsPrivate)
		assert.Equal(t, "extId1", resp.Responses[0].ExternalID)
		assert.Equal(t, "dashboard", resp.Responses[0].Type)
		assert.Equal(t, 1, resp.Responses[0].Version)
		assert.Equal(t, "owner1", resp.Responses[0].Owner)
		assert.Equal(t, http.StatusOK, resp.Responses[0].StatusCode)

		assert.Equal(t, "id2", resp.Responses[1].ID)
		assert.Equal(t, "name2", resp.Responses[1].Name)
		assert.Equal(t, false, resp.Responses[1].IsPrivate)
		assert.Equal(t, "extId2", resp.Responses[1].ExternalID)
		assert.Equal(t, "dashboard", resp.Responses[1].Type)
		assert.Equal(t, 1, resp.Responses[1].Version)
		assert.Equal(t, "owner2", resp.Responses[1].Owner)
		assert.Equal(t, http.StatusOK, resp.Responses[1].StatusCode)

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
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.List(ctx, "")

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
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.List(ctx, "")
		assert.Zero(t, resp)
		assert.Error(t, err)
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
	"externalId": "extId1",
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
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Delete(ctx, "")
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
						ContentType:  "multipart/form-data;boundary=Aas2UU1KdxSpaAyiNZ4-tnuzbwqnKuNK8vMOGy",
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
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Delete(ctx, "id-of-document")
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
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Delete(ctx, "id-of-document")

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
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Delete(ctx, "id-of-document")
		assert.Zero(t, resp)
		assert.Error(t, err)
	})
}
