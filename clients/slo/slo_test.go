// @license
// Copyright 2025 Dynatrace LLC
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

package slo_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/slo"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

func TestNewClient(t *testing.T) {
	actual := slo.NewClient(&rest.Client{})
	require.IsType(t, slo.Client{}, *actual)
}

func TestList(t *testing.T) {
	t.Run("successfully returns all SLOs", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "slos": [
    {
      "name": "CPU utilization",
      "description": "Measures the CPU usage of selected hosts over time.",
      "sliReference": "reference to slo template",
      "criteria": [
        {
          "timeframeFrom": "now-7d",
          "timeframeTo": "now",
          "target": 98
        }
      ],
      "tags": [],
      "id": "slo-id-1",
      "version": "ver1"
    }
  ]
}`
		apiResponse2 := `{
  "totalCount": 2,
  "slos": [
    {
      "name": "K8s namespace memory requests efficiency",
      "description": "Compares the actual usage of memory to the requested memory.",
      "customSli": "DQL based SLO",
      "criteria": [
        {
          "timeframeFrom": "now-7d",
          "timeframeTo": "now",
          "target": 80,
          "warning": 90
        }
      ],
      "tags": [],
      "id": "slo-id-2",
      "version": "ver1"
    }
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/slo/v1/slos", req.URL.Path)
					require.Equal(t, "", req.URL.Query().Get("page-key"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse1}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/slo/v1/slos", req.URL.Path)
					require.Equal(t, "key_for_next_page", req.URL.Query().Get("page-key"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse2}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 2, "for each call one listResponse should be present")
		assert.Len(t, resp.All(), 2, "two SLO objects in total should be downloaded")
	})

	t.Run("errors if can't execute all calls successfully", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "slos": [
    {
      "name": "CPU utilization",
      "description": "Measures the CPU usage of selected hosts over time.",
      "sliReference": "reference to slo template",
      "criteria": [
        {
          "timeframeFrom": "now-7d",
          "timeframeTo": "now",
          "target": 98
        }
      ],
      "tags": [],
      "id": "slo-id-1",
      "version": "ver1"
    }
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/slo/v1/slos", req.URL.Path)
					require.Equal(t, "", req.URL.Query().Get("page-key"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse1}
				},
			},
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusInternalServerError, ResponseBody: "Some error message from the server"}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "slo", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.List(t.Context())

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "slo", clientErr.Resource)
	})

	t.Run("errors if JSON unmarshaling fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/slo/v1/slos", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: "invalid json"}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "slo", runtimeErr.Resource)
		assert.Equal(t, "unmarshalling failed", runtimeErr.Reason)
	})
}

func TestGet(t *testing.T) {
	t.Run("successfully returns SLO for requested ID", func(t *testing.T) {
		getResponse := `{
  "name": "New SLO",
  "description": "This is a description",
  "customSli": {
    "indicator": "timeseries sli=avg(dt.host.cpu.idle)"
  },
  "criteria": [
    {
      "timeframeFrom": "now-7d",
      "timeframeTo": "now",
      "target": 99.8,
      "warning": 99.9
    }
  ],
  "tags": [
    "Stage:DEV"
  ],
  "id": "slo-id-1",
  "version": "ver1"
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/slo/v1/slos/slo-id-1", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: getResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "slo-id-1")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, getResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := slo.NewClient(&rest.Client{})

		actual, err := client.Get(t.Context(), "")

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "slo", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if SLO with ID doesn't exist on server", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 404,
    "message": "Provided ID 'false_ID' is not an SLO."
  }
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "false_ID")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "slo", clientErr.Resource)
		assert.Equal(t, "false_ID", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Get(t.Context(), "some-id")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "slo", clientErr.Resource)
		assert.Equal(t, "some-id", clientErr.Identifier)
	})
}

func TestCreate(t *testing.T) {
	t.Run("successfully creates a new SLO", func(t *testing.T) {
		given := `{
  "name": "CPU utilization",
  "description": "Measures the CPU usage of selected hosts over time.",
  "sliReference": "reference to slo template",
  "criteria": [
    {
      "timeframeFrom": "now-7d",
      "timeframeTo": "now",
      "target": 98
    }
  ],
  "tags": [],
  "externalId": "monaco_external_ID"
}`

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/slo/v1/slos", req.URL.Path)
					requestBody, _ := io.ReadAll(req.Body)
					require.JSONEq(t, given, string(requestBody))
					return testutils.Response{ResponseCode: http.StatusCreated}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.Create(t.Context(), []byte(given))

		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
		assert.Equal(t, http.StatusCreated, actual.StatusCode)
	})

	t.Run("errors if server returns an error", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 400,
    "message": "Invalid request body"
  }
}`

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusBadRequest, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.Create(t.Context(), []byte(`{}`))

		assert.Empty(t, actual)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "slo", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		actual, err := client.Create(t.Context(), []byte(`{}`))

		assert.Empty(t, actual)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "slo", clientErr.Resource)
	})
}

func TestUpdate(t *testing.T) {
	t.Run("successfully updates SLO with given payload", func(t *testing.T) {
		getResponse := `{
  "name": "CPU utilization",
  "description": "Measures the CPU usage of selected hosts over time.",
  "sliReference": "reference to slo template",
  "criteria": [
    {
      "timeframeFrom": "now-7d",
      "timeframeTo": "now",
      "target": 98
    }
  ],
  "tags": [],
  "id": "slo-id-1",
  "version": "ver1"
}`
		payload := `{
  "name": "K8s namespace memory requests efficiency",
  "description": "Compares the actual usage of memory to the requested memory.",
  "customSli": "DQL based SLO",
  "criteria": [
    {
      "timeframeFrom": "now-7d",
      "timeframeTo": "now",
      "target": 80,
      "warning": 90
    }
  ],
  "tags": [],
  "externalID": "monaco-external-ID"
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/slo/v1/slos/slo-id-1", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: getResponse}
				},
			},
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/slo/v1/slos/slo-id-1", req.URL.Path)
					require.Equal(t, "ver1", req.URL.Query().Get("optimistic-locking-version"))
					requestBody, _ := io.ReadAll(req.Body)
					require.JSONEq(t, payload, string(requestBody))
					return testutils.Response{ResponseCode: http.StatusOK}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "slo-id-1", []byte(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := slo.NewClient(&rest.Client{})

		actual, err := client.Update(t.Context(), "", nil)

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "slo", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if SLO with ID doesn't exist on server", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 404,
    "message": "Provided ID 'false_ID' is not an SLO."
  }
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "uid", nil)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "slo", clientErr.Resource)
		assert.Equal(t, "uid", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails on GET", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Update(t.Context(), "some-id", []byte(`{}`))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "slo", clientErr.Resource)
		assert.Equal(t, "some-id", clientErr.Identifier)
	})

	t.Run("errors if server returns error on PUT", func(t *testing.T) {
		getResponse := `{
  "name": "CPU utilization",
  "id": "slo-id-1",
  "version": "ver1"
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/slo/v1/slos/slo-id-1", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: getResponse}
				},
			},
			{
				PUT: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusBadGateway, ResponseBody: "bad gateway"}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "slo-id-1", []byte(`{}`))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Equal(t, "slo", clientErr.Resource)
		assert.Equal(t, "slo-id-1", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
	})
}

func TestDelete(t *testing.T) {
	t.Run("successfully deletes SLO with ID from server", func(t *testing.T) {
		getResponse := `{
  "name": "CPU utilization",
  "description": "Measures the CPU usage of selected hosts over time.",
  "sliReference": "reference to slo template",
  "criteria": [
    {
      "timeframeFrom": "now-7d",
      "timeframeTo": "now",
      "target": 98
    }
  ],
  "tags": [],
  "id": "slo-id-1",
  "version": "ver1"
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/slo/v1/slos/slo-id-1", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: getResponse}
				},
			},
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/slo/v1/slos/slo-id-1", req.URL.Path)
					require.Equal(t, "ver1", req.URL.Query().Get("optimistic-locking-version"))
					return testutils.Response{ResponseCode: http.StatusNoContent}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.Delete(t.Context(), "slo-id-1")

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, actual.StatusCode)
	})

	t.Run("errors if called without ID parameter", func(t *testing.T) {
		client := slo.NewClient(&rest.Client{})

		actual, err := client.Delete(t.Context(), "")

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "slo", Field: "id", Reason: "is empty"})
	})

	t.Run("errors if SLO with ID doesn't exist on server", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 404,
    "message": "Provided ID 'false_ID' is not an SLO."
  }
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.Delete(t.Context(), "uid")

		assert.Empty(t, actual)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "slo", clientErr.Resource)
		assert.Equal(t, "uid", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails on GET", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		actual, err := client.Delete(t.Context(), "some-id")

		assert.Empty(t, actual)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "slo", clientErr.Resource)
		assert.Equal(t, "some-id", clientErr.Identifier)
	})

	t.Run("errors if server returns error on DELETE", func(t *testing.T) {
		getResponse := `{
  "name": "CPU utilization",
  "id": "slo-id-1",
  "version": "ver1"
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/slo/v1/slos/slo-id-1", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: getResponse}
				},
			},
			{
				DELETE: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusForbidden, ResponseBody: `{"error":{"code":403,"message":"Not authorized."}}`}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := slo.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.Delete(t.Context(), "slo-id-1")

		assert.Empty(t, actual)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "slo", clientErr.Resource)
		assert.Equal(t, "slo-id-1", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
	})
}
