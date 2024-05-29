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

package openpipeline_test

import (
	"encoding/json"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/openpipeline"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

func TestOpenPipelineClient_Get(t *testing.T) {
	const payload = `{
	"id": "bizevents",
	"editable": true,
	"version": "1716904550612-4770deb9105b4a5293c1edbcc6bf4412"
}
`

	t.Run("Get - no ID given", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, "", openpipeline.GetOptions{})
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
						ContentType:  "application/json",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, "bizevents", openpipeline.GetOptions{})
		assert.Nil(t, err)
		assert.Equal(t, payload, string(resp.Data))
	})

	t.Run("GET - Unable to make HTTP call", func(t *testing.T) {

		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, "bizevents", openpipeline.GetOptions{})
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

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Get(ctx, "bizevents", openpipeline.GetOptions{})
		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusBadRequest, apiError.StatusCode)
	})
}

func TestOpenPipelineClient_List(t *testing.T) {
	const payload = `[
	{
		"id": "logs",
		"editable": true
	},
	{
		"id": "events",
		"editable": true
	},
	{
		"id": "bizevents",
		"editable": false
	}
]`

	t.Run("List - OK", func(t *testing.T) {
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

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.List(ctx, openpipeline.ListOptions{Editable: true})
		assert.Nil(t, err)
		assert.Len(t, resp, 2)
		assert.Contains(t, resp, openpipeline.ListResponse{Id: "logs", Editable: true})
		assert.Contains(t, resp, openpipeline.ListResponse{Id: "events", Editable: true})
		assert.NotContains(t, resp, openpipeline.ListResponse{Id: "bizevents", Editable: false})

		resp, err = client.List(ctx, openpipeline.ListOptions{Editable: false})
		assert.Nil(t, err)
		assert.Len(t, resp, 1)
		assert.NotContains(t, resp, openpipeline.ListResponse{Id: "logs", Editable: true})
		assert.NotContains(t, resp, openpipeline.ListResponse{Id: "events", Editable: true})
		assert.Contains(t, resp, openpipeline.ListResponse{Id: "bizevents", Editable: false})
	})

	t.Run("List - Unable to make HTTP call", func(t *testing.T) {

		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.List(ctx, openpipeline.ListOptions{})
		assert.Zero(t, resp)
		assert.Error(t, err)
	})

	t.Run("List - API Call returned with != 2xx", func(t *testing.T) {
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

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.List(ctx, openpipeline.ListOptions{})
		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusBadRequest, apiError.StatusCode)
	})
}

func TestOpenPipelineClient_GetAll(t *testing.T) {

	const payloadList = `[
	{
		"id": "logs",
		"editable": true
	},
	{
		"id": "events",
		"editable": true
	},
	{
		"id": "bizevents",
		"editable": false
	}
]`
	const payloadGet = `{
	"id": "bizevents",
	"editable": true,
	"version": "1716904550612-4770deb9105b4a5293c1edbcc6bf4412"
}
`

	t.Run("GET - OK", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payloadList,
						ContentType:  "application/json",
					}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payloadGet,
						ContentType:  "application/json",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.GetAll(ctx, openpipeline.GetAllOptions{})
		assert.Nil(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, payloadGet, string(resp[0].Data))
	})

	t.Run("GET - Unable to make HTTP call", func(t *testing.T) {

		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.GetAll(ctx, openpipeline.GetAllOptions{})
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

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.GetAll(ctx, openpipeline.GetAllOptions{})
		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusBadRequest, apiError.StatusCode)
	})
}

func TestOpenPipelineClient_Update(t *testing.T) {

	const getPayload = `{
	"id": "bizevents",
	"editable": true,
	"version": "1716904550612-4770deb9105b4a5293c1edbcc6bf4412"
}
`

	const putPayload = `{
	"id": "bizevents",
	"editable": true
}
`

	t.Run("PUT - no ID given", func(t *testing.T) {
		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "", []byte(putPayload), openpipeline.UpdateOptions{})
		assert.Zero(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("PUT - OK", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusAccepted,
						ResponseBody: getPayload,
						ContentType:  "application/json",
					}
				},
			},
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusAccepted,
						ResponseBody: "",
					}
				},
				ValidateRequest: func(t *testing.T, request *http.Request) {
					body, err := io.ReadAll(request.Body)
					assert.Nil(t, err)
					var m map[string]interface{}
					json.Unmarshal(body, &m)
					assert.Equal(t, "1716904550612-4770deb9105b4a5293c1edbcc6bf4412", m["version"])
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "bizevents", []byte(putPayload), openpipeline.UpdateOptions{})
		assert.Nil(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})

	t.Run("PUT - Unable to make HTTP call", func(t *testing.T) {

		responses := []testutils.ResponseDef{}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "", []byte(putPayload), openpipeline.UpdateOptions{})
		assert.Zero(t, resp)
		assert.Error(t, err)
	})

	t.Run("PUT - GET API Call returned with != 2xx", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
						ContentType:  "application/json",
					}
				},
			},
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusBadRequest,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "bizevents", []byte(putPayload), openpipeline.UpdateOptions{})
		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusNotFound, apiError.StatusCode)
	})

	t.Run("PUT - API Call returned with != 2xx", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusAccepted,
						ResponseBody: getPayload,
						ContentType:  "application/json",
					}
				},
			},
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusBadRequest,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))
		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "bizevents", []byte(putPayload), openpipeline.UpdateOptions{})
		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusBadRequest, apiError.StatusCode)
	})
}
