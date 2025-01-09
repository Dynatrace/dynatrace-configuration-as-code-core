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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

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

		ctx := testutils.ContextWithLogger(t)
		mockClient := slo.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			List(ctx, gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse1)),
			}, nil).
			Do(func(_ any, options rest.RequestOptions) {
				assert.Equal(t, url.Values{"page-key": []string{""}}, options.QueryParams)
			})
		mockClient.EXPECT().
			List(ctx, gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse2)),
			}, nil).
			Do(func(_ any, options rest.RequestOptions) {
				assert.Equal(t, url.Values{"page-key": {"key_for_next_page"}}, options.QueryParams)
			})

		fsClient := slo.NewTestClient(mockClient)
		resp, err := fsClient.List(ctx)

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

		ctx := testutils.ContextWithLogger(t)
		mockClient := slo.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			List(ctx, gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse1)),
			}, nil).
			Do(func(_ any, req rest.RequestOptions) {
				assert.Equal(t, url.Values{"page-key": {""}}, req.QueryParams)
			})
		mockClient.EXPECT().
			List(ctx, gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("Some error message from the server")),
			}, nil).
			Do(func(_ any, req rest.RequestOptions) {
				assert.Equal(t, url.Values{"page-key": {"key_for_next_page"}}, req.QueryParams)
			})

		fsClient := slo.NewTestClient(mockClient)
		resp, err := fsClient.List(ctx)

		assert.Error(t, err)
		assert.ErrorAs(t, err, new(api.APIError))
		assert.Empty(t, resp)
	})
}

func TestGet(t *testing.T) {
	t.Run("If SLO with ID doesn't exists on server returns error", func(t *testing.T) {
		apiResponse := `{
  "error": {
    "code": 400,
    "message": "Provided ID 'false_ID' is not an SLO."
  }
}`
		ctx := testutils.ContextWithLogger(t)
		mockClient := slo.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "uid", gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := slo.NewTestClient(mockClient)
		resp, err := fsClient.Get(ctx, "uid")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, apiResponse, string(apiErr.Body))
	})

	t.Run("successful request for requested ID", func(t *testing.T) {
		apiResponse := `{
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
		ctx := testutils.ContextWithLogger(t)
		mockClient := slo.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "uid", gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := slo.NewTestClient(mockClient)
		resp, err := fsClient.Get(ctx, "uid")

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, apiResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestCreate(t *testing.T) {
	given := `
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
	"externalId": "monaco_external_ID"
}`

	ctx := testutils.ContextWithLogger(t)
	mockClient := slo.NewMockclient(gomock.NewController(t))
	mockClient.EXPECT().
		Create(ctx, []byte(given), gomock.AssignableToTypeOf(rest.RequestOptions{})).
		Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader(given)),
		}, nil)

	fsClient := slo.NewTestClient(mockClient)
	actual, err := fsClient.Create(ctx, json.RawMessage(given))

	assert.NoError(t, err)
	assert.NotEmpty(t, actual)
	assert.Equal(t, http.StatusCreated, actual.StatusCode)
	assert.JSONEq(t, given, string(actual.Data))
}

func TestUpdate(t *testing.T) {
	t.Run("If SLO with ID doesn't exists on server returns error", func(t *testing.T) {
		apiResponse := `{
  "error": {
    "code": 400,
    "message": "Provided ID 'false_ID' is not an SLO."
  }
}`
		ctx := testutils.ContextWithLogger(t)
		mockClient := slo.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "uid", gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := slo.NewTestClient(mockClient)
		resp, err := fsClient.Update(ctx, "uid", json.RawMessage("{}"))

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, apiResponse, string(apiErr.Body))
	})

	t.Run("update works", func(t *testing.T) {
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

		ctx := testutils.ContextWithLogger(t)
		mockClient := slo.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "slo-id-1", gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(getResponse)),
			}, nil)
		mockClient.EXPECT().
			Update(ctx, "slo-id-1", "ver1", json.RawMessage(payload), gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{StatusCode: http.StatusOK}, nil)

		fsClient := slo.NewTestClient(mockClient)
		resp, err := fsClient.Update(ctx, "slo-id-1", json.RawMessage(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
	})
}

func TestDelete(t *testing.T) {
	t.Run("If SLO with ID doesn't exists on server returns error", func(t *testing.T) {
		apiResponse := `{
	 "error": {
	   "code": 404,
	   "message": "Segment not found",
	   "errorDetails": []
	 }
	}`
		ctx := testutils.ContextWithLogger(t)
		mockClient := slo.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Delete(ctx, "uid", gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := slo.NewTestClient(mockClient)
		resp, err := fsClient.Delete(ctx, "uid")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		errors.As(err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, apiResponse, string(apiErr.Body))
	})

	t.Run("successfully deleted entity with ID from server", func(t *testing.T) {

		ctx := testutils.ContextWithLogger(t)
		mockClient := slo.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Delete(ctx, "uid", gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusNoContent,
			}, nil)

		fsClient := slo.NewTestClient(mockClient)
		resp, err := fsClient.Delete(ctx, "uid")

		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusNoContent)
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
