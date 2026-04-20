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
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/openpipeline"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

func TestNewClient(t *testing.T) {
	actual := openpipeline.NewClient(&rest.Client{})
	require.IsType(t, openpipeline.Client{}, *actual)
}

func TestGet(t *testing.T) {
	const payload = `{
	"id": "bizevents",
	"editable": true,
	"version": "1716904550612-4770deb9105b4a5293c1edbcc6bf4412"
}`

	t.Run("successfully returns configuration for requested ID", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/openpipeline/v1/configurations/bizevents", req.URL.Path)
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

		resp, err := client.Get(t.Context(), "bizevents")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, payload, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := openpipeline.NewClient(&rest.Client{})

		resp, err := client.Get(t.Context(), "")

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "openpipeline", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if configuration with ID doesn't exist on server", func(t *testing.T) {
		errorResponse := `{"error":{"code":404,"message":"Not found"}}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
						ResponseBody: errorResponse,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "false_ID")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "openpipeline", clientErr.Resource)
		assert.Equal(t, "false_ID", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Get(t.Context(), "bizevents")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "openpipeline", clientErr.Resource)
		assert.Equal(t, "bizevents", clientErr.Identifier)
	})
}

func TestList(t *testing.T) {
	const payload = `[
	{"id": "logs", "editable": true},
	{"id": "events", "editable": true},
	{"id": "bizevents", "editable": false}
]`

	t.Run("successfully returns all configurations", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/openpipeline/v1/configurations", req.URL.Path)
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

		resp, err := client.List(t.Context())

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, payload, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if server returns an error", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusBadRequest,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "openpipeline", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.List(t.Context())

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "openpipeline", clientErr.Resource)
	})
}

func TestGetAll(t *testing.T) {
	const payloadList = `[
	{"id": "logs", "editable": true},
	{"id": "events", "editable": true},
	{"id": "bizevents", "editable": false}
]`
	const payloadGet1 = `{"id": "logs", "editable": true, "version": "ver1"}`
	const payloadGet2 = `{"id": "events", "editable": true, "version": "ver2"}`
	const payloadGet3 = `{"id": "bizevents", "editable": false, "version": "ver3"}`

	t.Run("successfully returns all configurations with full details", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/openpipeline/v1/configurations", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payloadList,
						ContentType:  "application/json",
					}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/openpipeline/v1/configurations/logs", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payloadGet1,
						ContentType:  "application/json",
					}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/openpipeline/v1/configurations/events", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payloadGet2,
						ContentType:  "application/json",
					}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/openpipeline/v1/configurations/bizevents", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payloadGet3,
						ContentType:  "application/json",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAll(t.Context())

		assert.NoError(t, err)
		assert.Len(t, resp, 3)
		assert.Equal(t, payloadGet1, string(resp[0].Data))
		assert.Equal(t, payloadGet2, string(resp[1].Data))
		assert.Equal(t, payloadGet3, string(resp[2].Data))
	})

	t.Run("errors if list call fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusBadRequest,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAll(t.Context())

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "openpipeline", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.GetAll(t.Context())

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "openpipeline", clientErr.Resource)
	})

	t.Run("errors if list response contains invalid JSON", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: "not valid json",
						ContentType:  "application/json",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAll(t.Context())

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "openpipeline", runtimeErr.Resource)
		assert.Equal(t, "unmarshalling failed", runtimeErr.Reason)
	})

	t.Run("errors if individual GET call fails", func(t *testing.T) {
		errorResponse := `{"error":{"code":400,"message":"Bad request"}}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payloadList,
						ContentType:  "application/json",
					}
				},
			},
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusBadRequest,
						ResponseBody: errorResponse,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAll(t.Context())

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "openpipeline", clientErr.Resource)
		assert.Equal(t, "logs", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	})
}

func TestUpdate(t *testing.T) {
	const getPayload = `{
	"id": "bizevents",
	"editable": true,
	"version": "1716904550612-4770deb9105b4a5293c1edbcc6bf4412",
	"updateToken": "my-update-token"
}`

	const putPayload = `{
	"id": "bizevents",
	"editable": true
}`

	t.Run("successfully updates configuration with given payload", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/openpipeline/v1/configurations/bizevents", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  "application/json",
					}
				},
			},
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/openpipeline/v1/configurations/bizevents", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusAccepted,
						ResponseBody: "",
					}
				},
				ValidateRequest: func(t *testing.T, request *http.Request) {
					body, err := io.ReadAll(request.Body)
					require.NoError(t, err)
					var m map[string]interface{}
					require.NoError(t, json.Unmarshal(body, &m))
					assert.Equal(t, "1716904550612-4770deb9105b4a5293c1edbcc6bf4412", m["version"])
					assert.Equal(t, "my-update-token", m["updateToken"])
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "bizevents", []byte(putPayload))

		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := openpipeline.NewClient(&rest.Client{})

		resp, err := client.Update(t.Context(), "", []byte(putPayload))

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "openpipeline", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if HTTP request fails on GET", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Update(t.Context(), "bizevents", []byte(putPayload))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "openpipeline", clientErr.Resource)
		assert.Equal(t, "bizevents", clientErr.Identifier)
	})

	t.Run("errors if GET returns server error", func(t *testing.T) {
		errorResponse := `{"error":{"code":404,"message":"Not found"}}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
						ResponseBody: errorResponse,
						ContentType:  "application/json",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "bizevents", []byte(putPayload))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "openpipeline", clientErr.Resource)
		assert.Equal(t, "bizevents", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if GET response data is invalid JSON", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ContentType:  "application/json",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "bizevents", []byte(putPayload))

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "openpipeline", runtimeErr.Resource)
		assert.Equal(t, "bizevents", runtimeErr.Identifier)
		assert.Equal(t, "failed to unmarshal GET response", runtimeErr.Reason)
	})

	t.Run("errors if request payload is invalid JSON", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  "application/json",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "bizevents", []byte(""))

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "openpipeline", runtimeErr.Resource)
		assert.Equal(t, "bizevents", runtimeErr.Identifier)
		assert.Equal(t, "failed to unmarshal request payload", runtimeErr.Reason)
	})

	t.Run("errors if server returns error on PUT", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  "application/json",
					}
				},
			},
			{
				PUT: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusBadRequest,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "bizevents", []byte(putPayload))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Equal(t, "openpipeline", clientErr.Resource)
		assert.Equal(t, "bizevents", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	})

	t.Run("retries on conflict and succeeds on second attempt", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  "application/json",
					}
				},
			},
			{
				PUT: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusConflict,
					}
				},
			},
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  "application/json",
					}
				},
			},
			{
				PUT: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusAccepted,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "bizevents", []byte(putPayload))

		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})

	t.Run("errors after exhausting all retry attempts on conflict", func(t *testing.T) {
		updateResponses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getPayload,
						ContentType:  "application/json",
					}
				},
			},
			{
				PUT: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusConflict,
					}
				},
			},
		}
		var responses []testutils.ResponseDef
		for range 10 {
			responses = append(responses, updateResponses...)
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := openpipeline.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "bizevents", []byte(putPayload))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Equal(t, "openpipeline", clientErr.Resource)
		assert.Equal(t, "bizevents", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusConflict, apiErr.StatusCode)
	})
}
