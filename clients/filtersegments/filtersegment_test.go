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

package filtersegments_test

import (
	"errors"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/filtersegments"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	t.Run("Get - no ID given", func(t *testing.T) {
		ctx := testutils.ContextWithLogger(t)
		fsClient := filtersegments.NewTestClient(filtersegments.NewMockclient(gomock.NewController(t)))
		resp, err := fsClient.Get(ctx, "")

		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "missing required id")
	})

	t.Run("Get - ID not found", func(t *testing.T) {
		apiResponse := `{
  "error": {
    "code": 404,
    "message": "Filter-segment not found",
    "errorDetails": []
  }
}`
		ctx := testutils.ContextWithLogger(t)
		mockClient := filtersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "uid", gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := filtersegments.NewTestClient(mockClient)
		resp, err := fsClient.Get(ctx, "uid")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		errors.As(err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, apiResponse, string(apiErr.Body))
	})

	t.Run("Get - found OK", func(t *testing.T) {
		apiResponse := `{
  "uid": "D82a1jdA23a",
  "name": "dev_environment",
  "description": "only includes data of the dev environment",
  "variables": {
    "type": "query",
    "value": "fetch logs | limit 1"
  },
  "isPublic": false,
  "owner": "john.doe",
  "includes": [
    {
      "filter": "here goes the filter",
      "dataObject": "logs"
    },
    {
      "filter": "here goes another filter",
      "dataObject": "events"
    }
  ],
  "version": 1
}`
		ctx := testutils.ContextWithLogger(t)
		mockClient := filtersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "uid", gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := filtersegments.NewTestClient(mockClient)
		resp, err := fsClient.Get(ctx, "uid")

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, apiResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestGetAll(t *testing.T) {

	t.Run("GetAll - fails", func(t *testing.T) {
		apiResponse := `{ "err" : "something went wrong" }`
		ctx := testutils.ContextWithLogger(t)
		mockClient := filtersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			List(ctx).
			Return(&http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := filtersegments.NewTestClient(mockClient)
		resp, err := fsClient.GetAll(ctx)

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		errors.As(err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, apiResponse, string(apiErr.Body))
	})

	t.Run("GetAll - getting individual object fails", func(t *testing.T) {
		apiResponse := `{
  "filterSegments": [
    {"uid": "pC7j2sEDzAQ"}
  ]
}
`
		apiResponse2 := `{ "err" : "something went wrong" }`
		ctx := testutils.ContextWithLogger(t)
		mockClient := filtersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			List(ctx).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)
		mockClient.EXPECT().
			Get(ctx, "pC7j2sEDzAQ", gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(apiResponse2)),
			}, nil)

		fsClient := filtersegments.NewTestClient(mockClient)
		resp, err := fsClient.GetAll(ctx)

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		errors.As(err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, apiResponse2, string(apiErr.Body))
	})

	t.Run("GetAll - OK", func(t *testing.T) {
		apiResponse := `{
  "filterSegments": [
    {"uid": "qW5qn449RsG"},
    {"uid": "pC7j2sEDzAQ"}
  ]
}
`
		apiResponse2 := `    {
      "uid": "qW5qn449RsG",
      "name": "dev_environment",
      "description": "only includes data of the dev environment",
      "variables": {"type": "query", "value": "fetch logs | limit 1"},
      "isPublic": false,
      "owner": "2f321c04-566e-4779-b576-3c033b8cd9e9",
      "allowedOperations": ["READ", "WRITE", "DELETE"],
      "version": 2
    }`
		apiResponse3 := `   {
      "uid": "pC7j2sEDzAQ",
      "name": "dev_environment",
      "description": "only includes data of the dev environment",
      "variables": {"type": "query", "value": "fetch logs | limit 1"},
      "isPublic": false,
      "owner": "2f321c04-566e-4779-b576-3c033b8cd9e9",
      "allowedOperations": ["READ", "WRITE", "DELETE"],
      "version": 1
    }`

		ctx := testutils.ContextWithLogger(t)
		mockClient := filtersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			List(ctx).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)
		mockClient.EXPECT().
			Get(ctx, "qW5qn449RsG", gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse2)),
			}, nil)
		mockClient.EXPECT().
			Get(ctx, "pC7j2sEDzAQ", gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse3)),
			}, nil)

		fsClient := filtersegments.NewTestClient(mockClient)
		resp, err := fsClient.GetAll(ctx)

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, apiResponse2, string(resp[0].Data))
		assert.Equal(t, apiResponse3, string(resp[1].Data))

	})
}

func Test_Upsert(t *testing.T) {

	t.Run("Create - Create OK", func(t *testing.T) {
		payload := `{
  "uid": "qW5qn449RsG",
  "name": "dev_environment",
  "description": "only includes data of the dev environment",
  "variables": {
    "type": "query",
    "value": "fetch logs | limit 1"
  },
  "isPublic": false,
  "owner": "2f321c04-566e-4779-b576-3c033b8cd9e9",
  "allowedOperations": [
    "READ",
    "WRITE",
    "DELETE"
  ],
  "includes": []
}`
		apiResponse := `{
  "uid": "oKZQWWV0FpR",
  "name": "dev_environment",
  "description": "only includes data of the dev environment",
  "variables": {
    "type": "query",
    "value": "fetch logs | limit 1"
  },
  "isPublic": false,
  "owner": "2f321c04-566e-4779-b576-3c033b8cd9e9",
  "includes": [],
  "version": 1
}`

		ctx := testutils.ContextWithLogger(t)
		mockClient := filtersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "uid", gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)
		mockClient.EXPECT().
			Create(ctx, []byte(payload)).
			Return(&http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := filtersegments.NewTestClient(mockClient)
		resp, err := fsClient.Upsert(ctx, "uid", []byte(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, apiResponse, string(resp.Data))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("Create - Update OK", func(t *testing.T) {
		payload := `{
  "uid": "qW5qn449RsG",
  "name": "dev_environment",
  "description": "only includes data of the dev environment",
  "variables": {
    "type": "query",
    "value": "fetch logs | limit 1"
  },
  "isPublic": false,
  "owner": "2f321c04-566e-4779-b576-3c033b8cd9e9",
  "allowedOperations": [
    "READ",
    "WRITE",
    "DELETE"
  ],
  "includes": []
}`

		apiExistingResource := `{
  "uid": "D82a1jdA23a",
  "name": "dev_environment",
  "description": "only includes data of the dev environment",
  "variables": {
    "type": "query",
    "value": "fetch logs | limit 1"
  },
  "isPublic": false,
  "owner": "john.doe",
  "includes": [
    {
      "filter": "here goes the filter",
      "dataObject": "logs"
    },
    {
      "filter": "here goes another filter",
      "dataObject": "events"
    }
  ],
  "version": 2
}`
		apiResponse := `{
  "uid": "oKZQWWV0FpR",
  "name": "dev_environment",
  "description": "only includes data of the dev environment",
  "variables": {
    "type": "query",
    "value": "fetch logs | limit 1"
  },
  "isPublic": false,
  "owner": "2f321c04-566e-4779-b576-3c033b8cd9e9",
  "includes": [],
  "version": 2
}`

		ctx := testutils.ContextWithLogger(t)
		mockClient := filtersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "uid", gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiExistingResource)),
			}, nil)
		mockClient.EXPECT().
			Update(ctx, "uid", []byte(payload), rest.RequestOptions{
				QueryParams: map[string][]string{
					"optimistic-locking-version": {"2"}},
			}).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := filtersegments.NewTestClient(mockClient)
		resp, err := fsClient.Upsert(ctx, "uid", []byte(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, apiResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

}

func Test_Delete(t *testing.T) {
	t.Run("Delete - no ID given", func(t *testing.T) {
		ctx := testutils.ContextWithLogger(t)
		fsClient := filtersegments.NewTestClient(filtersegments.NewMockclient(gomock.NewController(t)))
		resp, err := fsClient.Delete(ctx, "")

		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "missing required id")
	})

	t.Run("Delete - ID not found", func(t *testing.T) {
		apiResponse := `{
	 "error": {
	   "code": 404,
	   "message": "Filter-segment not found",
	   "errorDetails": []
	 }
	}`
		ctx := testutils.ContextWithLogger(t)
		mockClient := filtersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Delete(ctx, "uid").
			Return(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := filtersegments.NewTestClient(mockClient)
		resp, err := fsClient.Delete(ctx, "uid")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		errors.As(err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, apiResponse, string(apiErr.Body))
	})

	t.Run("Delete - OK", func(t *testing.T) {

		ctx := testutils.ContextWithLogger(t)
		mockClient := filtersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Delete(ctx, "uid").
			Return(&http.Response{
				StatusCode: http.StatusNoContent,
			}, nil)

		fsClient := filtersegments.NewTestClient(mockClient)
		resp, err := fsClient.Delete(ctx, "uid")

		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusNoContent)
	})

}
