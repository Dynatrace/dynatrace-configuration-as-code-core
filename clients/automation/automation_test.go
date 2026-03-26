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

package automation_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/automation"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

func TestNewClient(t *testing.T) {
	actual := automation.NewClient(&rest.Client{})
	require.IsType(t, automation.Client{}, *actual)
}

func TestGet(t *testing.T) {
	const payload = `{ "id" : "91cc8988-2223-404a-a3f5-5f1a839ecd45", "data" : "some-data1" }`

	t.Run("successfully returns workflow for requested ID", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/automation/v1/workflows/91cc8988-2223-404a-a3f5-5f1a839ecd45", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: payload}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, payload, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("successfully returns business calendar for requested ID", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/automation/v1/business-calendars/91cc8988-2223-404a-a3f5-5f1a839ecd45", req.URL.Path)
					require.Empty(t, req.URL.Query()["adminAccess"], "business calendars should not use adminAccess")
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: payload}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), automation.BusinessCalendars, "91cc8988-2223-404a-a3f5-5f1a839ecd45")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, payload, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := automation.NewClient(&rest.Client{})

		resp, err := client.Get(t.Context(), automation.Workflows, "")

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "automation-workflow", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if workflow with ID doesn't exist on server", func(t *testing.T) {
		errorResponse := `{"error":{"code":404,"message":"Not found"}}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), automation.Workflows, "false_ID")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "automation-workflow", clientErr.Resource)
		assert.Equal(t, "false_ID", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Get(t.Context(), automation.Workflows, "some-id")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "automation-workflow", clientErr.Resource)
		assert.Equal(t, "some-id", clientErr.Identifier)
	})

	t.Run("retries without adminAccess if forbidden for workflows", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "true", req.URL.Query().Get("adminAccess"))
					return testutils.Response{ResponseCode: http.StatusForbidden, ResponseBody: "{}"}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Empty(t, req.URL.Query()["adminAccess"])
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: payload}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, payload, string(resp.Data))
	})
}

func TestCreate(t *testing.T) {
	const payload = `{ "id" : "91cc8988-2223-404a-a3f5-5f1a839ecd45", "data" : "some-data1" }`

	t.Run("successfully creates a new workflow", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/automation/v1/workflows", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusCreated, ResponseBody: payload}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Create(t.Context(), automation.Workflows, []byte(payload))

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, payload, string(resp.Data))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("errors if server returns an error", func(t *testing.T) {
		errorResponse := `{"error":{"code":500,"message":"Internal Server Error"}}`

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusInternalServerError, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Create(t.Context(), automation.Workflows, []byte(payload))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "automation-workflow", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Create(t.Context(), automation.Workflows, []byte(payload))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "automation-workflow", clientErr.Resource)
	})
}

func TestUpdate(t *testing.T) {
	const payload = `{ "id" : "91cc8988-2223-404a-a3f5-5f1a839ecd45", "data" : "some-data1" }`

	t.Run("successfully updates a workflow", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/automation/v1/workflows/91cc8988-2223-404a-a3f5-5f1a839ecd45", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: payload}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte(payload))

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, payload, string(resp.Data))
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := automation.NewClient(&rest.Client{})

		resp, err := client.Update(t.Context(), automation.Workflows, "", []byte(payload))

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "automation-workflow", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if payload is invalid JSON", func(t *testing.T) {
		client := automation.NewClient(&rest.Client{})

		resp, err := client.Update(t.Context(), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte("invalid data"))

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "automation-workflow", runtimeErr.Resource)
		assert.Equal(t, "failed to remove id field from payload", runtimeErr.Reason)
	})

	t.Run("errors if server returns an error", func(t *testing.T) {
		errorResponse := `{"error":{"code":500,"message":"Internal Server Error"}}`

		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusInternalServerError, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte(payload))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Equal(t, "automation-workflow", clientErr.Resource)
		assert.Equal(t, "91cc8988-2223-404a-a3f5-5f1a839ecd45", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Update(t.Context(), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte(payload))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Equal(t, "automation-workflow", clientErr.Resource)
		assert.Equal(t, "91cc8988-2223-404a-a3f5-5f1a839ecd45", clientErr.Identifier)
	})

	t.Run("retries without adminAccess if forbidden for workflows", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "true", req.URL.Query().Get("adminAccess"))
					return testutils.Response{ResponseCode: http.StatusForbidden, ResponseBody: "{}"}
				},
			},
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Empty(t, req.URL.Query()["adminAccess"])
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: payload}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte(payload))

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, payload, string(resp.Data))
	})
}

func TestDelete(t *testing.T) {
	t.Run("successfully deletes workflow with ID from server", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/automation/v1/workflows/91cc8988-2223-404a-a3f5-5f1a839ecd45", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: "{}"}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Delete(t.Context(), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := automation.NewClient(&rest.Client{})

		resp, err := client.Delete(t.Context(), automation.Workflows, "")

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "automation-workflow", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if workflow with ID doesn't exist on server", func(t *testing.T) {
		errorResponse := `{"error":{"code":404,"message":"Not found"}}`

		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Delete(t.Context(), automation.Workflows, "false_ID")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "automation-workflow", clientErr.Resource)
		assert.Equal(t, "false_ID", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Delete(t.Context(), automation.Workflows, "some-id")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "automation-workflow", clientErr.Resource)
		assert.Equal(t, "some-id", clientErr.Identifier)
	})

	t.Run("uses adminAccess for workflow deletes", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "true", req.URL.Query().Get("adminAccess"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: "{}"}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Delete(t.Context(), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("does not use adminAccess for business calendar deletes", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Empty(t, req.URL.Query()["adminAccess"])
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: "{}"}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Delete(t.Context(), automation.BusinessCalendars, "91cc8988-2223-404a-a3f5-5f1a839ecd45")

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("retries without adminAccess if forbidden for workflows", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "true", req.URL.Query().Get("adminAccess"))
					return testutils.Response{ResponseCode: http.StatusForbidden, ResponseBody: "{}"}
				},
			},
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Empty(t, req.URL.Query()["adminAccess"])
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: "{}"}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Delete(t.Context(), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 2, server.Calls())
	})

	t.Run("errors if adminAccess forbidden and retry also fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusForbidden, ResponseBody: "{}"}
				},
			},
			{
				DELETE: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusInternalServerError, ResponseBody: "{}"}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Delete(t.Context(), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")

		assert.Empty(t, resp)
		assert.Equal(t, 2, server.Calls())

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "automation-workflow", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})
}

func TestList(t *testing.T) {
	t.Run("successfully returns all workflows across pages", func(t *testing.T) {
		apiResponse1 := `{ "count": 3, "results": [
			{"id": "82e7e7a4-dc69-4a7f-b0ad-7123f579ddf6", "title": "Workflow1"},
			{"id": "da105889-3817-435a-8b15-ec9777374b99", "title": "Workflow2"}
		]}`
		apiResponse2 := `{ "count": 3, "results": [
			{"id": "82e7e7a4-dc69-4a7f-b0ad-7123f579ddf6", "title": "Workflow3"}
		]}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "0", req.URL.Query().Get("offset"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse1}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "2", req.URL.Query().Get("offset"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse2}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context(), automation.Workflows)

		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Len(t, resp[0].Objects, 2)
		assert.Len(t, resp[1].Objects, 1)
	})

	t.Run("retries without adminAccess if forbidden for workflows", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "true", req.URL.Query().Get("adminAccess"))
					return testutils.Response{ResponseCode: http.StatusForbidden, ResponseBody: "{}"}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Empty(t, req.URL.Query()["adminAccess"])
					require.Equal(t, "0", req.URL.Query().Get("offset"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: `{ "count": 2, "results": [{"id": "id1", "title": "Workflow1"}]}`}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Empty(t, req.URL.Query()["adminAccess"])
					require.Equal(t, "1", req.URL.Query().Get("offset"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: `{ "count": 2, "results": [{"id": "id2", "title": "Workflow2"}]}`}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context(), automation.Workflows)

		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Len(t, resp[0].Objects, 1)
		assert.Len(t, resp[1].Objects, 1)
	})

	t.Run("errors if can't execute all calls successfully", func(t *testing.T) {
		apiResponse1 := `{ "count": 3, "results": [
			{"id": "82e7e7a4-dc69-4a7f-b0ad-7123f579ddf6", "title": "Workflow1"},
			{"id": "da105889-3817-435a-8b15-ec9777374b99", "title": "Workflow2"}
		]}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "0", req.URL.Query().Get("offset"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse1}
				},
			},
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusInternalServerError, ResponseBody: "{}"}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context(), automation.Workflows)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "automation-workflow", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.List(t.Context(), automation.Workflows)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "automation-workflow", clientErr.Resource)
	})

	t.Run("errors if server returns non-2xx", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusBadRequest, ResponseBody: "{}"}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context(), automation.Workflows)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "automation-workflow", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	})

	t.Run("errors if JSON unmarshaling fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: "invalid json"}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context(), automation.Workflows)

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "automation-workflow", runtimeErr.Resource)
		assert.Equal(t, "unmarshalling failed", runtimeErr.Reason)
	})
}

func TestListPaginationLogic(t *testing.T) {
	// prepare test data
	workflows := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		u, err := uuid.NewRandom()
		assert.NoError(t, err)
		workflows[i] = []byte(fmt.Sprintf(`{"id": "%s","title": "Workflow%d"}`, u, i))
	}

	responseTmpl := `{ "count": %d,"results": [ %s ] }`

	getResponse := func(t *testing.T, pageSize int, offsetQuery []string, serverData [][]byte) string {
		offset := 0
		if len(offsetQuery) > 0 {
			assert.Len(t, offsetQuery, 1)
			var err error
			offset, err = strconv.Atoi(offsetQuery[0])
			if err != nil {
				t.Fatalf("failed to parse query string: %v", err)
			}
		}

		end := offset + pageSize
		if end > len(serverData) {
			end = len(serverData)
		}

		var s []string
		for _, b := range serverData[offset:end] {
			s = append(s, string(b))
		}

		return fmt.Sprintf(responseTmpl, len(serverData), strings.Join(s, ","))
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		res := getResponse(t, 15, req.URL.Query()["offset"], workflows)
		_, _ = rw.Write([]byte(res))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	assert.NoError(t, err)

	client := automation.NewClient(rest.NewClient(u, server.Client()))

	resp, err := client.List(t.Context(), automation.Workflows)
	assert.NoError(t, err)

	assert.Len(t, resp, 7) // expect 7 pages - 6x full size 15, 1x size 10
	for i := 0; i < 6; i++ {
		assert.Len(t, resp[i].Objects, 15)
	}
	assert.Len(t, resp[6].Objects, 10)

	assert.ElementsMatch(t, workflows, resp.All())
}
