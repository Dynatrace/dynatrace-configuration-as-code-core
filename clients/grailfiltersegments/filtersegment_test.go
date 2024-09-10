// @license
// Copyright 2024 Dynatrace LLC
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

package grailfiltersegments_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/grailfiltersegments"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

func TestList(t *testing.T) {
	apiResponse := `{
  "filterSegments": [
    {
      "uid": "QElQbQcjq3S",
      "name": "filter_name",
      "isPublic": false,
      "owner": "userUUID",
      "version": 1
    }
  ]
}`
	expected := `[
    {
      "uid": "QElQbQcjq3S",
      "name": "filter_name",
      "isPublic": false,
      "owner": "userUUID",
      "version": 1
    }
  ]`

	mockClient := grailfiltersegments.NewMockclient(gomock.NewController(t))
	mockClient.EXPECT().
		List(gomock.Any(), rest.RequestOptions{}).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(apiResponse)),
		}, nil)

	fsClient := grailfiltersegments.NewTestClient(mockClient)
	actual, err := fsClient.List(context.Background())

	require.NoError(t, err)
	require.JSONEq(t, expected, string(actual.Data))
}

func TestGet(t *testing.T) {
	t.Run("call without ID returns an error", func(t *testing.T) {
		ctx := testutils.ContextWithLogger(t)
		fsClient := grailfiltersegments.NewTestClient(grailfiltersegments.NewMockclient(gomock.NewController(t)))
		resp, err := fsClient.Get(ctx, "")

		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "missing required id")
	})

	t.Run("ID doesn't exists on server returns error", func(t *testing.T) {
		apiResponse := `{
  "error": {
    "code": 404,
    "message": "Filter-segment not found",
    "errorDetails": []
  }
}`
		ctx := testutils.ContextWithLogger(t)
		mockClient := grailfiltersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "uid", gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := grailfiltersegments.NewTestClient(mockClient)
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
		mockClient := grailfiltersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "uid", gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := grailfiltersegments.NewTestClient(mockClient)
		resp, err := fsClient.Get(ctx, "uid")

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, apiResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestGetAll(t *testing.T) {
	t.Run("getting list fails with error", func(t *testing.T) {
		apiResponse := `{ "err" : "something went wrong" }`
		ctx := testutils.ContextWithLogger(t)
		mockClient := grailfiltersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			List(ctx, rest.RequestOptions{}).
			Return(&http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := grailfiltersegments.NewTestClient(mockClient)
		resp, err := fsClient.GetAll(ctx)

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		errors.As(err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, apiResponse, string(apiErr.Body))
	})

	t.Run("getting individual object from server fails and return error", func(t *testing.T) {
		apiResponse := `{
  "filterSegments": [
    {"uid": "pC7j2sEDzAQ"}
  ]
}
`
		apiResponse2 := `{ "err" : "something went wrong" }`
		ctx := testutils.ContextWithLogger(t)
		mockClient := grailfiltersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			List(ctx, rest.RequestOptions{}).
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

		fsClient := grailfiltersegments.NewTestClient(mockClient)
		resp, err := fsClient.GetAll(ctx)

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		errors.As(err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, apiResponse2, string(apiErr.Body))
	})

	t.Run("successfully returned all configuration from server", func(t *testing.T) {
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
		mockClient := grailfiltersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			List(ctx, rest.RequestOptions{}).
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

		fsClient := grailfiltersegments.NewTestClient(mockClient)
		resp, err := fsClient.GetAll(ctx)

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, apiResponse2, string(resp[0].Data))
		assert.Equal(t, apiResponse3, string(resp[1].Data))

	})
}

func TestUpsert(t *testing.T) {

	t.Run("successfully created new entity on server", func(t *testing.T) {
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
		mockClient := grailfiltersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "uid", gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)
		mockClient.EXPECT().
			Create(ctx, []byte(payload), rest.RequestOptions{}).
			Return(&http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := grailfiltersegments.NewTestClient(mockClient)
		resp, err := fsClient.Upsert(ctx, "uid", []byte(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, apiResponse, string(resp.Data))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("successfully updated existing entity on server", func(t *testing.T) {
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
		mockClient := grailfiltersegments.NewMockclient(gomock.NewController(t))
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

		fsClient := grailfiltersegments.NewTestClient(mockClient)
		resp, err := fsClient.Upsert(ctx, "uid", []byte(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, apiResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

}

func TestDelete(t *testing.T) {
	t.Run("call without ID returns an error", func(t *testing.T) {
		ctx := testutils.ContextWithLogger(t)
		fsClient := grailfiltersegments.NewTestClient(grailfiltersegments.NewMockclient(gomock.NewController(t)))
		resp, err := fsClient.Delete(ctx, "")

		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "missing required id")
	})

	t.Run("ID doesn't exists on server returns error", func(t *testing.T) {
		apiResponse := `{
	 "error": {
	   "code": 404,
	   "message": "Filter-segment not found",
	   "errorDetails": []
	 }
	}`
		ctx := testutils.ContextWithLogger(t)
		mockClient := grailfiltersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Delete(ctx, "uid", rest.RequestOptions{}).
			Return(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := grailfiltersegments.NewTestClient(mockClient)
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
		mockClient := grailfiltersegments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Delete(ctx, "uid", rest.RequestOptions{}).
			Return(&http.Response{
				StatusCode: http.StatusNoContent,
			}, nil)

		fsClient := grailfiltersegments.NewTestClient(mockClient)
		resp, err := fsClient.Delete(ctx, "uid")

		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusNoContent)
	})
}
