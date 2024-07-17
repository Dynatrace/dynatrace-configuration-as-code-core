/*
 * @license
 * Copyright 2023 Dynatrace LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package documents_test

import (
	"net/http"
	"testing"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/documents"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentClient_Create(t *testing.T) {
	const payload = `{
    "id": "038ab74f-0a3a-4bf8-9068-85e2d633a1e6",
    "name": "my-test-db",
	"isPrivate": true,
	"externalId": "extId",
    "type": "dashboard",
    "version": 1
}`

	t.Run("OK", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusCreated,
						ResponseBody: payload,
					}
				},
				ValidateRequest: func(t *testing.T, request *http.Request) {
					require.NotNil(t, request.Header.Get("ContentType"))
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		given := documents.Document{
			Kind:       "type",
			Name:       "my_name",
			ExternalID: "some ID",
			Content:    []byte("the content can be anything"),
		}
		resp, err := client.Create(ctx, given)
		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
	})

	t.Run("Unable to make HTTP POST call", func(t *testing.T) {

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
		given := documents.Document{
			Kind:       "type",
			Name:       "my_name",
			ExternalID: "some ID",
			Content:    []byte("can be anything"),
		}
		resp, err := client.Create(ctx, given)
		assert.Zero(t, resp)
		assert.Error(t, err)
	})

	t.Run("API Call returned with != 2xx - No change", func(t *testing.T) {

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
		given := documents.Document{
			Kind:       "type",
			Name:       "my-dashboard",
			ExternalID: "extId",
			Content:    []byte(payload),
		}

		resp, err := client.Create(ctx, given)

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
	})
}

func TestDocumentClient_Patch(t *testing.T) {
	const patchPayload = `{
  "documentMetadata": {
    "id": "038ab74f-0a3a-4bf8-9068-85e2d633a1e6",
    "name": "my-test-db",
	"isPrivate": true,
	"externalId": "extId",
    "type": "dashboard",
    "version": 1,
    "owner": "12341234-1234-1234-1234-12341234"
  }
}`

	t.Run("OK", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				PATCH: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: patchPayload,
					}
				},
				ValidateRequest: func(t *testing.T, request *http.Request) {
					assert.Contains(t, request.URL.Path, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6")
					assert.Contains(t, request.URL.Query(), "optimistic-locking-version")
					assert.Equal(t, "1", request.URL.Query().Get("optimistic-locking-version"))
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()
		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Patch(ctx, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "1", documents.Document{})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
	})

	t.Run("Missing id", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Patch(ctx, "", "1", documents.Document{})
		assert.Zero(t, resp)
		assert.ErrorContains(t, err, "id")
	})

	t.Run("Missing version", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Patch(ctx, "someID", "", documents.Document{})
		assert.Zero(t, resp)
		assert.ErrorContains(t, err, "version")
	})

	t.Run("Document not found", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				PATCH: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
					}
				},
				ValidateRequest: func(t *testing.T, request *http.Request) {
					assert.Contains(t, request.URL.Path, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6")
					assert.Contains(t, request.URL.Query(), "optimistic-locking-version")
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := documents.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Patch(ctx, "038ab74f-0a3a-4bf8-9068-85e2d633a1e6", "1", documents.Document{})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
	})
}
