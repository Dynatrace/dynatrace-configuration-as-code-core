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

package automation_test

import (
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/automation"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/internal/testutils"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestAutomationClient_Get(t *testing.T) {
	const payload = `{ "id" : "91cc8988-2223-404a-a3f5-5f1a839ecd45", "data" : "some-data1" }`

	t.Run("Get - no ID given", func(t *testing.T) {
		responses := testutils.ServerResponses{}
		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, automation.Workflows, "")
		assert.Zero(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("GET - OK", func(t *testing.T) {

		responses := testutils.ServerResponses{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: payload,
			},
		}
		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
		assert.NotNil(t, resp)
		assert.Equal(t, payload, string(resp.Data))
		assert.NoError(t, err)
	})

	t.Run("GET - Unable to make HTTP call", func(t *testing.T) {
		responses := testutils.ServerResponses{}
		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
		assert.Zero(t, resp)
		assert.Error(t, err)
	})

	t.Run("GET - API Call returned with != 2xx", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodGet: {
				ResponseCode: http.StatusBadRequest,
			},
		}
		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.NoError(t, err)
	})
}

func TestAutomationClient_Upsert(t *testing.T) {

	_ = `{"id" : "91cc8988-2223-404a-a3f5-5f1a839ecd45", "data" : "some-data"}`

	t.Run("Upsert - no ID given", func(t *testing.T) {
		responses := testutils.ServerResponses{}
		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Upsert(ctx, automation.Workflows, "", []byte{})
		assert.Zero(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("Upsert - invalid data", func(t *testing.T) {
		responses := testutils.ServerResponses{}
		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Upsert(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte{})
		assert.Zero(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("Upsert - not able to make HTTP PUT call", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodPut: {
				ResponseCode: http.StatusBadRequest,
				ResponseBody: "{}",
			},
		}

		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))
		ctx := testutils.ContextWithLogger(t)
		data := []byte(`{"id" : "some-id"}`)
		resp, err := client.Upsert(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", data)
		assert.Zero(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("Upsert - adminAccess query parameter set for workflows", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodPut: {
				ResponseCode: http.StatusOK,
				ResponseBody: "{}",
				ValidateRequestFunc: func(request *http.Request) {
					adminAccessQP := request.URL.Query()["adminAccess"]
					assert.Len(t, adminAccessQP, 1)
					assert.Equal(t, "true", adminAccessQP[0])

				},
			},
		},
			{
				http.MethodPut: {
					ResponseCode: http.StatusOK,
					ResponseBody: "{}",
					ValidateRequestFunc: func(request *http.Request) {
						adminAccessQP := request.URL.Query()["adminAccess"]
						assert.Len(t, adminAccessQP, 1)
						assert.Equal(t, "false", adminAccessQP[0])
					},
				},
			}}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		client.Upsert(testutils.ContextWithLogger(t), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte(`{"id" : "some-id"}`))
		client.Upsert(testutils.ContextWithLogger(t), automation.BusinessCalendars, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte(`{"id" : "some-id"}`))
	})

	t.Run("Upsert - adminAccess forbidden", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodPut: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "{}",
				ValidateRequestFunc: func(request *http.Request) {
					adminAccessQP := request.URL.Query()["adminAccess"]
					assert.Len(t, adminAccessQP, 1)
					assert.Equal(t, "true", adminAccessQP[0])
				},
			},
		}, {
			http.MethodPut: {
				ResponseCode: http.StatusOK,
				ResponseBody: "{}",
				ValidateRequestFunc: func(request *http.Request) {
					adminAccessQP := request.URL.Query()["adminAccess"]
					assert.Len(t, adminAccessQP, 0)
				},
			},
		}}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		data := []byte(`{"id" : "some-id"}`)
		resp, err := client.Upsert(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", data)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 2, server.Calls())
		assert.Nil(t, err)
	})

	t.Run("Upsert - Direct update using HTTP PUT API Call returned with != 2xx- creation via POST fails", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodPut: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "{}",
			},
		}, {
			http.MethodPut: {
				ResponseCode: http.StatusNotFound,
				ResponseBody: "{}",
			},
		}, {
			http.MethodPost: {
				ResponseCode: http.StatusInternalServerError,
				ResponseBody: "{}",
			},
		}}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		data := []byte(`{"id" : "some-id"}`)
		resp, err := client.Upsert(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", data)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Equal(t, 3, server.Calls())
		assert.Nil(t, err)
	})

	t.Run("Upsert - Direct update using HTTP PUT API Call returned with != 2xx - creation via POST OK", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodPut: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "{}",
			},
		}, {
			http.MethodPut: {
				ResponseCode: http.StatusNotFound,
				ResponseBody: "{}",
			},
		}, {
			http.MethodPost: {
				ResponseCode: http.StatusCreated,
				ResponseBody: "{}",
			},
		}}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		data := []byte(`{"id" : "some-id"}`)
		resp, err := client.Upsert(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", data)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, 3, server.Calls())
		assert.Nil(t, err)
	})
}

func TestAutomationClient_Delete(t *testing.T) {
	t.Run("Delete - no ID given", func(t *testing.T) {
		responses := testutils.ServerResponses{}
		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Delete(ctx, automation.Workflows, "")
		assert.Zero(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("Delete - HTTP Call fails", func(t *testing.T) {
		responses := testutils.ServerResponses{}
		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Delete(ctx, automation.Workflows, "")
		assert.Zero(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("Delete - adminAccess query parameter set for workflows", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodDelete: {
				ResponseCode: http.StatusOK,
				ResponseBody: "{}",
				ValidateRequestFunc: func(request *http.Request) {
					adminAccessQP := request.URL.Query()["adminAccess"]
					assert.Len(t, adminAccessQP, 1)
					assert.Equal(t, "true", adminAccessQP[0])

				},
			},
		},
			{
				http.MethodDelete: {
					ResponseCode: http.StatusOK,
					ResponseBody: "{}",
					ValidateRequestFunc: func(request *http.Request) {
						adminAccessQP := request.URL.Query()["adminAccess"]
						assert.Len(t, adminAccessQP, 1)
						assert.Equal(t, "false", adminAccessQP[0])
					},
				},
			}}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		client.Delete(testutils.ContextWithLogger(t), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
		client.Delete(testutils.ContextWithLogger(t), automation.BusinessCalendars, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
	})

	t.Run("Delete - adminAccess forbidden", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodDelete: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "{}",
				ValidateRequestFunc: func(request *http.Request) {
					adminAccessQP := request.URL.Query()["adminAccess"]
					assert.Len(t, adminAccessQP, 1)
					assert.Equal(t, "true", adminAccessQP[0])
				},
			},
		}, {
			http.MethodDelete: {
				ResponseCode: http.StatusOK,
				ResponseBody: "{}",
				ValidateRequestFunc: func(request *http.Request) {
					adminAccessQP := request.URL.Query()["adminAccess"]
					assert.Len(t, adminAccessQP, 0)
				},
			},
		}}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Delete(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 2, server.Calls())
		assert.Nil(t, err)
	})

	t.Run("Delete - adminAccess forbidden - DELETE API Call returned with != 2xx", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodDelete: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "{}",
			},
		}, {
			http.MethodDelete: {
				ResponseCode: http.StatusInternalServerError,
				ResponseBody: "{}",
			},
		}}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Delete(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Equal(t, 2, server.Calls())
		assert.Nil(t, err)
	})

	t.Run("Delete - adminAccess forbidden - resource not found", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodDelete: {
				ResponseCode: http.StatusNotFound,
				ResponseBody: "{}",
			},
		}}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Delete(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Equal(t, 1, server.Calls())
		assert.Nil(t, err)
	})
}

func TestAutomationClient_List(t *testing.T) {

	payloadePages := []string{`{ "count": 3,"results":
			[
				{"id": "82e7e7a4-dc69-4a7f-b0ad-7123f579ddf6","title": "Workflow1"},
				{"id": "da105889-3817-435a-8b15-ec9777374b99","title": "Workflow2"}
  			]
		}`,
		`{ "count": 3,"results":
			[
				{"id": "82e7e7a4-dc69-4a7f-b0ad-7123f579ddf6","title": "Workflow3"}
  			]
		}`,
	}

	t.Run("List - Paginated - OK", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: payloadePages[0],
				ValidateRequestFunc: func(request *http.Request) {
					assert.Equal(t, []string{"0"}, request.URL.Query()["offset"])
				},
			},
		},
			{
				http.MethodGet: {
					ResponseCode: http.StatusOK,
					ResponseBody: payloadePages[1],
					ValidateRequestFunc: func(request *http.Request) {
						assert.Equal(t, []string{"1"}, request.URL.Query()["offset"])
					},
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.List(ctx, automation.Workflows)
		assert.Len(t, resp, 2)
		assert.Len(t, resp[0].Objects, 2)
		assert.Len(t, resp[1].Objects, 1)
		assert.Nil(t, err)
	})

	t.Run("List - Paginated - Getting one page fails", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: payloadePages[0],
				ValidateRequestFunc: func(request *http.Request) {
					assert.Equal(t, []string{"0"}, request.URL.Query()["offset"])
				},
			},
		},
			{
				http.MethodGet: {
					ResponseCode: http.StatusInternalServerError,
					ResponseBody: "{}",
					ValidateRequestFunc: func(request *http.Request) {
						assert.Equal(t, []string{"1"}, request.URL.Query()["offset"])
					},
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.List(ctx, automation.Workflows)
		assert.Len(t, resp, 1)
		assert.Len(t, resp[0].Objects, 0)
		assert.Nil(t, err)
	})

	t.Run("List - API Call returned with != 2xx", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusBadRequest,
				ResponseBody: "{}",
			},
		},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.List(ctx, automation.Workflows)
		assert.Len(t, resp, 1)
		assert.Equal(t, resp[0].Response.StatusCode, http.StatusBadRequest)
		assert.Len(t, resp[0].Objects, 0)
		assert.Nil(t, err)
	})

	t.Run("List - API Call failed", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodGet: {},
		},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.List(ctx, automation.Workflows)
		assert.Equal(t, automation.PagedListResponse{}, resp)
		assert.NotNil(t, err)
	})
}

func TestPagedListResponse(t *testing.T) {
	pr := automation.PagedListResponse{
		automation.ListResponse{
			api.ListResponse{
				Response: api.Response{},
				Objects: [][]byte{
					{'1'},
					{'2'},
				},
			},
		},
		automation.ListResponse{
			api.ListResponse{
				Response: api.Response{},
				Objects: [][]byte{
					{'3'},
					{'4'},
				},
			},
		},
	}

	assert.Equal(t, [][]byte{{'1'}, {'2'}, {'3'}, {'4'}}, pr.Objects())
}
