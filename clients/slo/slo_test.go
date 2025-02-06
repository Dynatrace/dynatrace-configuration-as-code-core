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
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/slo"
)

func TestNewClient(t *testing.T) {
	actual := slo.NewClient(&rest.Client{})
	require.IsType(t, slo.Client{}, *actual)
}

func TestList(t *testing.T) {
	t.Run("successfully returned all configuration from server", func(t *testing.T) {
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
      "name": "K8s namespace memory requests efficiency - OLD",
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

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			require.Equal(t, http.MethodGet, request.Method)
			require.Equal(t, "/platform/slo/v1/slos", request.URL.Path)

			switch v := request.URL.Query().Get("page-key"); v {
			case "":
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(apiResponse1))
			case "key_for_next_page":
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(apiResponse2))
			default:
				require.Failf(t, "unexpected call", "unexpected call with page-key= %s", v)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := slo.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.List(context.TODO())

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 2, "for each call one listResponse should be present")
		assert.Len(t, resp.All(), 2, "two SLO object it total should be downloaded")
	})

	t.Run("Fails if can't execute all calls successfully", func(t *testing.T) {
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

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			require.Equal(t, http.MethodGet, request.Method)
			require.Equal(t, "/platform/slo/v1/slos", request.URL.Path)

			switch v := request.URL.Query().Get("page-key"); v {
			case "":
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(apiResponse1))
			case "key_for_next_page": // provoke error
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte("Some error message from the server"))
			default:
				require.Failf(t, "unexpected call", "unexpected call with page-key= %s", v)
			}
		}))

		url, _ := url.Parse(server.URL)
		client := slo.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.List(context.TODO())

		assert.Error(t, err)
		assert.ErrorAs(t, err, new(api.APIError))
		assert.Empty(t, resp)
	})
}

func TestGet(t *testing.T) {
	t.Run("when called without id parameter, returns an error", func(t *testing.T) {
		client := slo.NewClient(&rest.Client{})

		actual, err := client.Get(context.TODO(), "")

		assert.Error(t, err)
		assert.Empty(t, actual)
	})

	t.Run("successful request for requested ID", func(t *testing.T) {
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
  "id": "vu9U3hXa3q0AAAABAClidWlsdGluOmludGVybmFsLnNlcnZpY2UubGV2ZWwub2JqZWN0aXZlcwAGdGVuYW50AAZ0ZW5hbnQAJGZlMDZjZTBlLWNlM2QtMzY5OC1hNDI3LTA3OTU2Zjk0MTcyN77vVN4V2t6t",
  "version": "vu9U3hXY3q0ATAAkZmUwNmNlMGUtY2UzZC0zNjk4LWE0MjctMDc5NTZmOTQxNzI3ACRiOTlhYjNlMS1kNDFiLTExZWYtODAwMS1kNjlhZGNiNDJmOWa-71TeFdjerQ"
}`
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			t.Log(request.URL.String())
			require.Equal(t, http.MethodGet, request.Method)
			require.Equal(t, "/platform/slo/v1/slos/uid", request.URL.Path)

			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte(getResponse))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := slo.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.Get(context.TODO(), "uid")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, getResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("If SLO with ID doesn't exists on server returns error", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 400,
    "message": "Provided ID 'false_ID' is not an SLO."
  }
}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte(errorResponse))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := slo.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.Get(context.TODO(), "uid")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})
}

func TestCreate(t *testing.T) {
	given := `
{
   "name":"CPU utilization",
   "description":"Measures the CPU usage of selected hosts over time.",
   "sliReference":"reference to slo template",
   "criteria":[
      {
         "timeframeFrom":"now-7d",
         "timeframeTo":"now",
         "target":98
      }
   ],
   "tags":[ ],
   "externalId":"monaco_external_ID"
}`

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		t.Log(request.URL.String())
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, "/platform/slo/v1/slos", request.URL.Path)
		requestBody, _ := io.ReadAll(request.Body)
		require.JSONEq(t, given, string(requestBody))

		writer.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	url, _ := url.Parse(server.URL)
	client := slo.NewClient(rest.NewClient(url, server.Client()))

	actual, err := client.Create(context.TODO(), json.RawMessage(given))

	assert.NoError(t, err)
	assert.NotEmpty(t, actual)
	assert.Equal(t, http.StatusCreated, actual.StatusCode)
}

func TestUpdate(t *testing.T) {
	t.Run("when called without id parameter, returns an error", func(t *testing.T) {
		client := slo.NewClient(&rest.Client{})

		actual, err := client.Update(context.TODO(), "", nil)

		assert.Error(t, err)
		assert.Empty(t, actual)
	})

	t.Run("successful update of SLO with given payload", func(t *testing.T) {
		getResponse := `
{
   "name":"CPU utilization",
   "description":"Measures the CPU usage of selected hosts over time.",
   "sliReference":"reference to slo template",
   "criteria":[
      {
         "timeframeFrom":"now-7d",
         "timeframeTo":"now",
         "target":98
      }
   ],
   "tags":[ ],
   "id":"slo-id-1",
   "version":"ver1"
}`
		payload := `{
  "name": "K8s namespace memory requests efficiency - OLD",
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
  "externalID" : "monaco-external-ID"
}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodGet:
				require.Equal(t, "/platform/slo/v1/slos/slo-id-1", request.URL.Path)

				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(getResponse))
			case http.MethodPut:
				require.Equal(t, "/platform/slo/v1/slos/slo-id-1", request.URL.Path)
				require.Equal(t, "ver1", request.URL.Query().Get("optimistic-locking-version"), "'optimistic-locking-version' should be the same as 'version' value from the get response")
				requestBody, _ := io.ReadAll(request.Body)
				require.JSONEq(t, payload, string(requestBody))

				writer.WriteHeader(http.StatusOK)
			default:
				require.Failf(t, "unexpected http call", "unexpected http call: %s %s", request.Method, request.URL)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := slo.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.Update(context.TODO(), "slo-id-1", json.RawMessage(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
	})

	t.Run("update for non existing SLO fails with error", func(t *testing.T) {
		get404Response := `{
  "error": {
    "code": 404,
    "message": "Provided ID 'false_ID' is not an SLO."
  }
}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte(get404Response))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := slo.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.Update(context.TODO(), "uid", nil)

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, get404Response, string(apiErr.Body))
	})
}

func TestDelete(t *testing.T) {
	t.Run("when called without id parameter, returns an error", func(t *testing.T) {
		client := slo.NewClient(&rest.Client{})

		actual, err := client.Delete(context.TODO(), "")

		assert.Error(t, err)
		assert.Empty(t, actual)
	})

	t.Run("successfully deleted entity with ID from server", func(t *testing.T) {
		getResponse := `
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
}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodGet:
				require.Equal(t, "/platform/slo/v1/slos/slo-id-1", request.URL.Path)

				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(getResponse))
			case http.MethodDelete:
				require.Equal(t, "/platform/slo/v1/slos/slo-id-1", request.URL.Path)
				require.Equal(t, "ver1", request.URL.Query().Get("optimistic-locking-version"), "'optimistic-locking-version' should be the same as 'version' value from the get response")

				writer.WriteHeader(http.StatusNoContent)
			default:
				require.Failf(t, "unexpected http call", "unexpected http call: %s %s", request.Method, request.URL)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := slo.NewClient(rest.NewClient(url, server.Client()))

		actual, err := client.Delete(context.TODO(), "slo-id-1")

		assert.NoError(t, err)
		assert.Equal(t, actual.StatusCode, http.StatusNoContent)
	})

	t.Run("If SLO with ID doesn't exists on server returns an error", func(t *testing.T) {
		get404Response := `{
  "error": {
    "code": 404,
    "message": "Provided ID 'false_ID' is not an SLO."
  }
}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte(get404Response))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := slo.NewClient(rest.NewClient(url, server.Client()))

		actual, err := client.Delete(context.TODO(), "uid")

		assert.Empty(t, actual)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		errors.As(err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, get404Response, string(apiErr.Body))
	})
}

func TestUnmarshallFromListResponse(t *testing.T) {
	given := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "slos": [
{ "data": "fist" },
{ "data": "second" },
{ "data": "third" }
]
}`
	expected := [][]byte{
		[]byte(`{ "data": "fist" }`),
		[]byte(`{ "data": "second" }`),
		[]byte(`{ "data": "third" }`),
	}

	nextPage, actual, err := slo.UnmarshallFromListResponse([]byte(given))
	assert.NoError(t, err)
	assert.Equal(t, "key_for_next_page", nextPage)
	assert.NotEmpty(t, actual)
	assert.Equal(t, expected, actual)
}
