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

package segments_test

import (
	"context"
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
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/segments"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

func TestNewClient(t *testing.T) {
	actual := segments.NewClient(&rest.Client{})
	require.IsType(t, segments.Client{}, *actual)
}

func TestList(t *testing.T) {
	apiResponse := `{
  "filterSegments": [
    {
      "uid": "QElQbQcjq3S",
      "name": "segment_name",
      "isPublic": false,
      "owner": "userUUID",
      "version": 1
    }
  ]
}`
	expected := `[
    {
      "uid": "QElQbQcjq3S",
      "name": "segment_name",
      "isPublic": false,
      "owner": "userUUID",
      "version": 1
    }
  ]`

	mockClient := segments.NewMockclient(gomock.NewController(t))
	mockClient.EXPECT().
		List(gomock.AssignableToTypeOf(context.Background()), gomock.AssignableToTypeOf(rest.RequestOptions{})).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(apiResponse)),
		}, nil)

	fsClient := segments.NewTestClient(mockClient)
	actual, err := fsClient.List(context.Background())

	require.NoError(t, err)
	require.JSONEq(t, expected, string(actual.Data))
}

func TestGet(t *testing.T) {
	t.Run("ID doesn't exists on server returns error", func(t *testing.T) {
		apiResponse := `{
  "error": {
    "code": 404,
    "message": "Segment not found",
    "errorDetails": []
  }
}`
		ctx := testutils.ContextWithLogger(t)
		mockClient := segments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "uid", gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := segments.NewTestClient(mockClient)
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
		mockClient := segments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "uid", gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := segments.NewTestClient(mockClient)
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
		mockClient := segments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			List(ctx, gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := segments.NewTestClient(mockClient)
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
		mockClient := segments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			List(ctx, gomock.AssignableToTypeOf(rest.RequestOptions{})).
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

		fsClient := segments.NewTestClient(mockClient)
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
		mockClient := segments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			List(ctx, gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)
		mockClient.EXPECT().
			Get(ctx, "qW5qn449RsG", gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse2)),
			}, nil)
		mockClient.EXPECT().
			Get(ctx, "pC7j2sEDzAQ", gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse3)),
			}, nil)

		fsClient := segments.NewTestClient(mockClient)
		resp, err := fsClient.GetAll(ctx)

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, apiResponse2, string(resp[0].Data))
		assert.Equal(t, apiResponse3, string(resp[1].Data))

	})
}

func TestCreate(t *testing.T) {
	t.Run("successfully created new entity on server", func(t *testing.T) {
		payload := `{
  "name": "dev_environment",
  "description": "only includes data of the dev environment",
  "variables": {
    "type": "query",
    "value": "fetch logs | limit 1"
  },
  "isPublic": false,
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
		mockClient := segments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Create(ctx, []byte(payload), gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := segments.NewTestClient(mockClient)
		resp, err := fsClient.Create(ctx, []byte(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, apiResponse, string(resp.Data))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

func TestUpdate(t *testing.T) {

	t.Run("unexpected error while checking for status on server", func(t *testing.T) {
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
		ctx := testutils.ContextWithLogger(t)
		mockClient := segments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "uid", gomock.Any()).
			Return(nil, errors.New("some unexpected error"))

		fsClient := segments.NewTestClient(mockClient)
		actual, err := fsClient.Update(ctx, "uid", []byte(payload))

		require.Error(t, err)
		require.Empty(t, actual)
	})

	t.Run("successfully updated existing entity on server, uid in provided payload not matching and gets overwritten", func(t *testing.T) {
		uid := "D82a1jdA23a"
		payload := `{
  "allowedOperations" : [ "READ", "WRITE", "DELETE" ],
  "description" : "only includes data of the dev environment",
  "includes" : [ ],
  "isPublic" : false,
  "name" : "dev_environment",
  "owner" : "2f321c04-566e-4779-b576-3c033b8cd9e9",
  "uid" : "uid",
  "variables" : {
    "type" : "query",
    "value" : "fetch logs | limit 1"
  }
}`

		apiExistingResource := `{
  "uid": "` + uid + `",
  "name": "dev_environment",
  "description": "only includes data of the dev environment",
  "variables": {
    "type": "query",
    "value": "fetch logs | limit 1"
  },
  "isPublic": false,
  "owner": "2f321c04-566e-4779-b576-3c033b8cd9e9",
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
		apiResponse := `{}`

		ctx := testutils.ContextWithLogger(t)
		mockClient := segments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, uid, gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiExistingResource)),
			}, nil)
		mockClient.EXPECT().
			Update(ctx, uid, gomock.Any(), gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil).
			Do(func(_, _, payload any, ro any) {
				assertVersionParam(t, ro, "2")
				assertRequestPayload(payload, t, uid, "2f321c04-566e-4779-b576-3c033b8cd9e9")
			})

		fsClient := segments.NewTestClient(mockClient)
		resp, err := fsClient.Update(ctx, uid, []byte(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, apiResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("successfully updated existing entity on server, payload without owner and uid", func(t *testing.T) {
		payload := `{
  "allowedOperations" : [ "READ", "WRITE", "DELETE" ],
  "description" : "only includes data of the dev environment",
  "includes" : [ ],
  "isPublic" : false,
  "name" : "dev_environment",
  "variables" : {
    "type" : "query",
    "value" : "fetch logs | limit 1"
  }
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
  "owner": "2f321c04-566e-4779-b576-3c033b8cd9e9",
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
		apiResponse := `{}`

		ctx := testutils.ContextWithLogger(t)
		mockClient := segments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Get(ctx, "D82a1jdA23a", gomock.Any()).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiExistingResource)),
			}, nil)
		mockClient.EXPECT().
			Update(ctx, "D82a1jdA23a", gomock.Any(), gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil).
			Do(func(_, _, payload any, ro any) {
				assertVersionParam(t, ro, "2")
				assertRequestPayload(payload, t, "D82a1jdA23a", "2f321c04-566e-4779-b576-3c033b8cd9e9")
			})

		fsClient := segments.NewTestClient(mockClient)
		resp, err := fsClient.Update(ctx, "D82a1jdA23a", []byte(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, apiResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestDelete(t *testing.T) {
	t.Run("ID doesn't exists on server returns error", func(t *testing.T) {
		apiResponse := `{
	 "error": {
	   "code": 404,
	   "message": "Segment not found",
	   "errorDetails": []
	 }
	}`
		ctx := testutils.ContextWithLogger(t)
		mockClient := segments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Delete(ctx, "uid", gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(apiResponse)),
			}, nil)

		fsClient := segments.NewTestClient(mockClient)
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
		mockClient := segments.NewMockclient(gomock.NewController(t))
		mockClient.EXPECT().
			Delete(ctx, "uid", gomock.AssignableToTypeOf(rest.RequestOptions{})).
			Return(&http.Response{
				StatusCode: http.StatusNoContent,
			}, nil)

		fsClient := segments.NewTestClient(mockClient)
		resp, err := fsClient.Delete(ctx, "uid")

		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusNoContent)
	})
}

func assertVersionParam(t *testing.T, ro any, version string) {
	require.IsType(t, rest.RequestOptions{}, ro)
	require.Contains(t, ro.(rest.RequestOptions).QueryParams, "optimistic-locking-version")
	require.Equal(t, ro.(rest.RequestOptions).QueryParams, url.Values{"optimistic-locking-version": {version}})
}

func assertRequestPayload(payload any, t *testing.T, expectedUID string, expectedOwner string) {
	data, ok := payload.([]byte)
	if !ok {
		t.Error("invalid payload type")
	}
	var testRequest struct {
		Owner string `json:"owner"`
		UID   string `json:"uid"`
	}
	err := json.Unmarshal(data, &testRequest)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, testRequest.Owner, expectedOwner)
	assert.Equal(t, testRequest.UID, expectedUID)
}
