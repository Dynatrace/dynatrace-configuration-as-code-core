package documents_test

import (
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/documents"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
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
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.NoError(t, err)
	})
}

func TestDocumentClient_Create(t *testing.T) {
	const payload = `{
    "modificationInfo": {
        "createdBy": "12341234-1234-1234-1234-12341234",
        "createdTime": "2024-04-11T14:06:26.491Z",
        "lastModifiedBy": "2f321c04-566e-4779-b576-3c033b8cd9e9",
        "lastModifiedTime": "2024-04-11T14:06:26.491Z"
    },
    "access": [
        "read",
        "delete",
        "write"
    ],
    "id": "038ab74f-0a3a-4bf8-9068-85e2d633a1e6",
    "name": "my-test-db",
    "type": "dashboard",
    "version": 1,
    "owner": "12341234-1234-1234-1234-12341234"
}`

	t.Run("Create  - OK", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusCreated,
						ResponseBody: payload,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Create(ctx, "my-dashboard", []byte(payload), documents.Dashboard)
		assert.NotNil(t, resp)
		assert.Equal(t, payload, string(resp.Data))
		assert.NoError(t, err)
	})

	t.Run("Create - API Call returned with != 2xx", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
						ResponseBody: "{}"}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Create(ctx, "my-dashboard", []byte(payload), documents.Dashboard)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, err)
	})

	t.Run("Create - Unable to make HTTP POST call", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusCreated,
						ResponseBody: payload,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Create(ctx, "my-dashboard", []byte(payload), documents.Dashboard)
		assert.Zero(t, resp)
		assert.Error(t, err)
	})
}

func TestDocumentClient_Upsert(t *testing.T) {
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
	const payload = `{
    "modificationInfo": {
        "createdBy": "12341234-1234-1234-1234-12341234",
        "createdTime": "2024-04-11T14:06:26.491Z",
        "lastModifiedBy": "2f321c04-566e-4779-b576-3c033b8cd9e9",
        "lastModifiedTime": "2024-04-11T14:06:26.491Z"
    },
    "access": [
        "read",
        "delete",
        "write"
    ],
    "id": "038ab74f-0a3a-4bf8-9068-85e2d633a1e6",
    "name": "my-test-db",
    "type": "dashboard",
    "version": 1,
    "owner": "12341234-1234-1234-1234-12341234"
}`

	const patchPayload = `{
  "documentMetadata": {
    "modificationInfo": {
      "createdBy": "12341234-1234-1234-1234-12341234",
      "createdTime": "2024-04-11T14:06:26.491Z",
      "lastModifiedBy": "2f321c04-566e-4779-b576-3c033b8cd9e9",
      "lastModifiedTime": "2024-04-11T14:06:26.491Z"
    },
    "access": [
      "read",
      "delete",
      "write"
    ],
    "id": "038ab74f-0a3a-4bf8-9068-85e2d633a1e6",
    "name": "my-test-db",
    "type": "dashboard",
    "version": 1,
    "owner": "12341234-1234-1234-1234-12341234"
  }
}`

	t.Run("Upsert - Missing id", func(t *testing.T) {
		responses := []testutils.ResponseDef{}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Upsert(ctx, "", "my-dashboard", []byte(payload), documents.Dashboard)
		assert.Zero(t, resp)
		assert.Error(t, err)
	})

	t.Run("Upsert - No document found - Creates new Document  - OK", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
					}
				},
			},
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusCreated,
						ResponseBody: payload,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Upsert(ctx, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", []byte(payload), documents.Dashboard)
		assert.NotNil(t, resp)
		assert.Equal(t, payload, string(resp.Data))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.NoError(t, err)
	})

	t.Run("Upsert - Fails to fetch existing document", func(t *testing.T) {
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
		resp, err := client.Upsert(ctx, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", []byte(payload), documents.Dashboard)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, err)
	})

	t.Run("Upsert - Existing Document Found - Updates it", func(t *testing.T) {
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
		resp, err := client.Upsert(ctx, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", []byte(payload), documents.Dashboard)
		assert.NoError(t, err)
		assert.Equal(t, patchPayload, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		fmt.Println(resp.ID)
	})

	t.Run("Upsert - Existing Document Found - Update fails", func(t *testing.T) {
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
		resp, err := client.Upsert(ctx, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "my-dashboard", []byte(payload), documents.Dashboard)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
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
		assert.Equal(t, "dashboard", resp.Responses[0].Type)
		assert.Equal(t, 1, resp.Responses[0].Version)
		assert.Equal(t, "owner1", resp.Responses[0].Owner)
		assert.Equal(t, http.StatusOK, resp.Responses[0].StatusCode)

		assert.Equal(t, "id2", resp.Responses[1].ID)
		assert.Equal(t, "name2", resp.Responses[1].Name)
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
		assert.Len(t, resp.Responses, 0)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, err)
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
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  "multipart/form-data;boundary=Aas2UU1KdxSpaAyiNZ4-tnuzbwqnKuNK8vMOGy",
					}
				},
			},
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
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
		assert.NotZero(t, resp)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.NoError(t, err)
	})

	t.Run("Delete - Failed to execut Request", func(t *testing.T) {
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
