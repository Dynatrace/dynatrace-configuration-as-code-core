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
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/automation"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
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

func TestAutomationClient_Create(t *testing.T) {
	const payload = `{ "id" : "91cc8988-2223-404a-a3f5-5f1a839ecd45", "data" : "some-data1" }`
	t.Run("Create  - OK", func(t *testing.T) {

		responses := testutils.ServerResponses{
			http.MethodPost: {
				ResponseCode: http.StatusCreated,
				ResponseBody: payload,
			},
		}
		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Create(ctx, automation.Workflows, []byte(payload))
		assert.NotNil(t, resp)
		assert.Equal(t, payload, string(resp.Data))
		assert.NoError(t, err)
	})

	t.Run("Create - HTTP PUT returns non 2xx", func(t *testing.T) {

		responses := testutils.ServerResponses{
			http.MethodPost: {
				ResponseCode: http.StatusInternalServerError,
				ResponseBody: "{}",
			},
		}
		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Create(ctx, automation.Workflows, []byte(payload))

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, err)
	})

	t.Run("Create - Unable to make HTTP POST call", func(t *testing.T) {

		responses := testutils.ServerResponses{
			http.MethodPost: {
				ResponseCode: http.StatusCreated,
				ResponseBody: payload,
			},
		}
		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Create(ctx, automation.Workflows, []byte(payload))
		assert.Zero(t, resp)
		assert.Error(t, err)
	})
}

func TestAutomationClient_Update(t *testing.T) {
	const payload = `{ "id" : "91cc8988-2223-404a-a3f5-5f1a839ecd45", "data" : "some-data1" }`
	t.Run("Update  - try with adminAccess -if fails try without - OK", func(t *testing.T) {

		responses := []testutils.ServerResponses{{
			http.MethodPut: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: payload,
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
					ResponseBody: payload,
					ValidateRequestFunc: func(request *http.Request) {
						adminAccessQP := request.URL.Query()["adminAccess"]
						assert.Len(t, adminAccessQP, 0)
					},
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte(payload))
		assert.NotNil(t, resp)
		assert.Equal(t, payload, string(resp.Data))
		assert.NoError(t, err)
	})

	t.Run("Update - HTTP PUT returns non 2xx", func(t *testing.T) {

		responses := testutils.ServerResponses{
			http.MethodPut: {
				ResponseCode: http.StatusInternalServerError,
				ResponseBody: "{}",
			},
		}
		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte(payload))

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, err)
	})

	t.Run("Update - HTTP PUT call is not possible", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodPut: {
				ResponseCode: http.StatusInternalServerError,
				ResponseBody: "{}",
			},
		}
		server := testutils.NewHTTPTestServer(t, []testutils.ServerResponses{responses})
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte(payload))

		assert.Zero(t, resp)
		assert.Error(t, err)
	})
}

func TestAutomationClient_Upsert(t *testing.T) {

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
						assert.Equal(t, "true", adminAccessQP[0])
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

	t.Run("Delete - adminAccess query parameter set", func(t *testing.T) {
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
						assert.Equal(t, "true", adminAccessQP[0])
					},
				},
			}}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		_, _ = client.Delete(testutils.ContextWithLogger(t), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
		_, _ = client.Delete(testutils.ContextWithLogger(t), automation.BusinessCalendars, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
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
						assert.Equal(t, []string{"2"}, request.URL.Query()["offset"])
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

	t.Run("List - Paginated - With Admin Permissions Missing", func(t *testing.T) {
		responses := []testutils.ServerResponses{
			{
				http.MethodGet: {
					ResponseCode: http.StatusForbidden,
					ResponseBody: "{}",
					ValidateRequestFunc: func(request *http.Request) {
						assert.Equal(t, []string{"true"}, request.URL.Query()["adminAccess"])
					},
				},
			},
			{
				http.MethodGet: {
					ResponseCode: http.StatusOK,
					ResponseBody: `{ "count": 2,"results": [ {"id": "82e7e7a4-dc69-4a7f-b0ad-7123f579ddf6","title": "Workflow1"} ] }`,
					ValidateRequestFunc: func(request *http.Request) {
						assert.Equal(t, []string{"false"}, request.URL.Query()["adminAccess"])
						assert.Equal(t, []string{"0"}, request.URL.Query()["offset"])
					},
				},
			},
			{
				http.MethodGet: {
					ResponseCode: http.StatusOK,
					ResponseBody: `{ "count": 2,"results": [ {"id": "da105889-3817-435a-8b15-ec9777374b99","title": "Workflow2"} ] }`,
					ValidateRequestFunc: func(request *http.Request) {
						assert.Equal(t, []string{"false"}, request.URL.Query()["adminAccess"])
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
		assert.Len(t, resp[0].Objects, 1)
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
						assert.Equal(t, []string{"2"}, request.URL.Query()["offset"])
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
		assert.Equal(t, api.PagedListResponse{}, resp)
		assert.NotNil(t, err)
	})
}

func TestAutomationClient_List_PaginationLogic(t *testing.T) {

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

	ctx := testutils.ContextWithLogger(t)
	resp, err := client.List(ctx, automation.Workflows)
	assert.Nil(t, err)

	assert.Len(t, resp, 7) // expect 7 pages - 6x full size 15, 1x size 10
	for i := 0; i < 6; i++ {
		assert.Len(t, resp[i].Objects, 15)
	}
	assert.Len(t, resp[6].Objects, 10)

	assert.ElementsMatch(t, workflows, resp.All())
}
