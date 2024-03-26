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
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/automation"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

func TestAutomationClient_Get(t *testing.T) {
	const payload = `{ "id" : "91cc8988-2223-404a-a3f5-5f1a839ecd45", "data" : "some-data1" }`

	t.Run("Get - no ID given", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, automation.Workflows, "")
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
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
		assert.NotNil(t, resp)
		body, _ := io.ReadAll(resp.Body)
		defer resp.Body.Close()

		assert.Equal(t, payload, string(body))
		assert.NoError(t, err)
	})

	t.Run("GET - Unable to make HTTP call", func(t *testing.T) {

		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
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

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Create(ctx, automation.Workflows, []byte(payload))
		assert.NotNil(t, resp)
		body, _ := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		assert.Equal(t, payload, string(body))
		assert.NoError(t, err)
	})

	t.Run("Create - HTTP PUT returns non 2xx", func(t *testing.T) {

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

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Create(ctx, automation.Workflows, []byte(payload))

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

		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusForbidden,
						ResponseBody: payload,
					}
				},
				ValidateRequest: func(t *testing.T, req *http.Request) {
					adminAccessQP := req.URL.Query()["adminAccess"]
					assert.Len(t, adminAccessQP, 1)
					assert.Equal(t, "true", adminAccessQP[0])
				},
			},
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payload,
					}
				},
				ValidateRequest: func(t *testing.T, req *http.Request) {
					adminAccessQP := req.URL.Query()["adminAccess"]
					assert.Len(t, adminAccessQP, 0)
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte(payload))
		assert.NotNil(t, resp)
		body, _ := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		assert.Equal(t, payload, string(body))
		assert.NoError(t, err)
	})

	t.Run("Update - HTTP PUT returns non 2xx", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
						ResponseBody: "{}",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte(payload))

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, err)
	})

	t.Run("Update - HTTP PUT call is not possible", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
						ResponseBody: "{}",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45", []byte(payload))

		assert.Zero(t, resp)
		assert.Error(t, err)
	})
}

func TestAutomationClient_Delete(t *testing.T) {
	t.Run("Delete - no ID given", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Delete(ctx, automation.Workflows, "")
		assert.Zero(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("Delete - HTTP Call fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Delete(ctx, automation.Workflows, "")
		assert.Zero(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("Delete - adminAccess query parameter set", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: "{}",
					}
				},
				ValidateRequest: func(t *testing.T, req *http.Request) {
					adminAccessQP := req.URL.Query()["adminAccess"]
					assert.Len(t, adminAccessQP, 1)
					assert.Equal(t, "true", adminAccessQP[0])

				},
			},
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: "{}",
					}
				},
				ValidateRequest: func(t *testing.T, req *http.Request) {
					adminAccessQP := req.URL.Query()["adminAccess"]
					assert.Nil(t, adminAccessQP)
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		_, _ = client.Delete(testutils.ContextWithLogger(t), automation.Workflows, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
		_, _ = client.Delete(testutils.ContextWithLogger(t), automation.BusinessCalendars, "91cc8988-2223-404a-a3f5-5f1a839ecd45")
	})

	t.Run("Delete - adminAccess forbidden", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusForbidden,
						ResponseBody: "{}",
					}
				},
				ValidateRequest: func(t *testing.T, req *http.Request) {
					adminAccessQP := req.URL.Query()["adminAccess"]
					assert.Len(t, adminAccessQP, 1)
					assert.Equal(t, "true", adminAccessQP[0])
				},
			},
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: "{}",
					}
				},
				ValidateRequest: func(t *testing.T, req *http.Request) {
					adminAccessQP := req.URL.Query()["adminAccess"]
					assert.Len(t, adminAccessQP, 0)
				},
			},
		}

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

		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusForbidden,
						ResponseBody: "{}",
					}
				},
			},
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
						ResponseBody: "{}",
					}
				},
			},
		}

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

		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
						ResponseBody: "{}",
					}
				},
			},
		}

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
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payloadePages[0],
					}
				},
				ValidateRequest: func(t *testing.T, req *http.Request) {
					assert.Equal(t, []string{"0"}, req.URL.Query()["offset"])
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payloadePages[1],
					}
				},
				ValidateRequest: func(t *testing.T, req *http.Request) {
					assert.Equal(t, []string{"2"}, req.URL.Query()["offset"])
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.List(ctx, automation.Workflows, 0)
		assert.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, payloadePages[0], string(body))
		resp, err = client.List(ctx, automation.Workflows, 2)
		body, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.NoError(t, err)
		assert.Equal(t, payloadePages[1], string(body))
	})

	t.Run("List - API Call returned with != 2xx", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusBadRequest,
						ResponseBody: "{}",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.List(ctx, automation.Workflows, 0)
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusBadRequest)
	})

	t.Run("List - API Call failed", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := automation.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.List(ctx, automation.Workflows, 0)
		assert.Nil(t, resp)
		assert.NotNil(t, err)
	})
}
