// @license
// Copyright 2026 Dynatrace LLC
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

package extensions_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/extensions"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

func TestNewClient(t *testing.T) {
	actual := extensions.NewClient(&rest.Client{})
	require.IsType(t, extensions.Client{}, *actual)
}

const nextPageKeyParam = "next-page-key"
const pageSizeParam = "page-size"

func TestListExtensions(t *testing.T) {
	t.Run("successfully returns all extensions across multiple pages", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "items": [
    {"extensionName": "com.dynatrace.extension.foo", "version": "1.0.0"}
  ]
}`
		apiResponse2 := `{
  "totalCount": 2,
  "items": [
    {"extensionName": "com.dynatrace.extension.bar", "version": "2.0.0"}
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/extensions/v2/extensions", req.URL.Path)
					require.Equal(t, "100", req.URL.Query().Get(pageSizeParam))
					require.Equal(t, "", req.URL.Query().Get(nextPageKeyParam))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse1}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/extensions/v2/extensions", req.URL.Path)
					require.Equal(t, "key_for_next_page", req.URL.Query().Get(nextPageKeyParam))
					require.Equal(t, "", req.URL.Query().Get(pageSizeParam))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse2}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.ListExtensions(t.Context())

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Len(t, resp, 2, "for each call one listResponse should be present")
		assert.Len(t, resp.All(), 2, "two extension objects in total should be downloaded")
	})

	t.Run("errors if can't execute all calls successfully", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "items": [
    {"extensionName": "com.dynatrace.extension.foo", "version": "1.0.0"}
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
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

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.ListExtensions(t.Context())

		assert.Empty(t, resp)
		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.ListExtensions(t.Context())

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.ClientError{})
	})

	t.Run("errors if JSON unmarshalling fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: `invalid json`}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.ListExtensions(t.Context())

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.RuntimeError{})
	})
}

func TestListExtensionVersions(t *testing.T) {
	t.Run("successfully returns all versions across multiple pages", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "pageSize": 100,
  "items": [
    {"extensionName": "com.dynatrace.extension.foo", "version": "1.0.0"}
  ]
}`
		apiResponse2 := `{
  "totalCount": 2,
  "pageSize": 100,
  "items": [
    {"extensionName": "com.dynatrace.extension.foo", "version": "1.1.0"}
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/extensions/v2/extensions/com.dynatrace.extension.foo", req.URL.Path)
					require.Equal(t, "100", req.URL.Query().Get(pageSizeParam))
					require.Equal(t, "", req.URL.Query().Get(nextPageKeyParam))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse1}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/extensions/v2/extensions/com.dynatrace.extension.foo", req.URL.Path)
					require.Equal(t, "key_for_next_page", req.URL.Query().Get(nextPageKeyParam))
					require.Equal(t, "", req.URL.Query().Get(pageSizeParam))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse2}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.ListExtensionVersions(t.Context(), "com.dynatrace.extension.foo")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Len(t, resp, 2, "for each call one listResponse should be present")
		assert.Len(t, resp.All(), 2, "two version objects in total should be downloaded")
	})

	t.Run("errors if called without extension name", func(t *testing.T) {
		client := extensions.NewClient(&rest.Client{})

		resp, err := client.ListExtensionVersions(t.Context(), "")

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "extensions", Field: "extension-name", Reason: "is empty"})
	})

	t.Run("errors if can't execute all calls successfully", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "pageSize": 100,
  "items": [
    {"extensionName": "com.dynatrace.extension.foo", "version": "1.0.0"}
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
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

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.ListExtensionVersions(t.Context(), "com.dynatrace.extension.foo")

		assert.Empty(t, resp)
		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.ListExtensionVersions(t.Context(), "com.dynatrace.extension.foo")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.ClientError{})
	})

	t.Run("errors if JSON unmarshalling fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: `invalid json`}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.ListExtensionVersions(t.Context(), "com.dynatrace.extension.foo")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.RuntimeError{})
	})
}

func TestListMonitoringConfigurations(t *testing.T) {
	t.Run("successfully returns all monitoring configurations across multiple pages", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "items": [
    {"objectId": "config-id-1", "value": {"enabled": true}}
  ]
}`
		apiResponse2 := `{
  "totalCount": 2,
  "items": [
    {"objectId": "config-id-2", "value": {"enabled": false}}
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/extensions/v2/extensions/com.dynatrace.extension.foo/monitoring-configurations", req.URL.Path)
					require.Equal(t, "500", req.URL.Query().Get(pageSizeParam))
					require.Equal(t, "", req.URL.Query().Get(nextPageKeyParam))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse1}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/extensions/v2/extensions/com.dynatrace.extension.foo/monitoring-configurations", req.URL.Path)
					require.Equal(t, "key_for_next_page", req.URL.Query().Get(nextPageKeyParam))
					require.Equal(t, "", req.URL.Query().Get(pageSizeParam))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: apiResponse2}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.ListMonitoringConfigurations(t.Context(), "com.dynatrace.extension.foo")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Len(t, resp, 2, "for each call one listResponse should be present")
		assert.Len(t, resp.All(), 2, "two monitoring configuration objects in total should be downloaded")
	})

	t.Run("errors if called without extension name", func(t *testing.T) {
		client := extensions.NewClient(&rest.Client{})

		resp, err := client.ListMonitoringConfigurations(t.Context(), "")

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "extensions", Field: "extension-name", Reason: "is empty"})
	})

	t.Run("errors if can't execute all calls successfully", func(t *testing.T) {
		apiResponse1 := `{
  "totalCount": 2,
  "nextPageKey": "key_for_next_page",
  "items": [
    {"objectId": "config-id-1", "value": {"enabled": true}}
  ]
}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
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

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.ListMonitoringConfigurations(t.Context(), "com.dynatrace.extension.foo")

		assert.Empty(t, resp)
		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.ListMonitoringConfigurations(t.Context(), "com.dynatrace.extension.foo")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.ClientError{})
	})

	t.Run("errors if JSON unmarshalling fails", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: `invalid json`}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.ListMonitoringConfigurations(t.Context(), "com.dynatrace.extension.foo")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.RuntimeError{})
	})
}

func TestGetEnvironmentConfiguration(t *testing.T) {
	t.Run("successfully returns environment configuration", func(t *testing.T) {
		getResponse := `{"version": "1.2.3"}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/extensions/v2/extensions/com.dynatrace.extension.foo/environment-configuration", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: getResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetEnvironmentConfiguration(t.Context(), "com.dynatrace.extension.foo")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, getResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without extension name", func(t *testing.T) {
		client := extensions.NewClient(&rest.Client{})

		resp, err := client.GetEnvironmentConfiguration(t.Context(), "")

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "extensions", Field: "extension-name", Reason: "is empty"})
	})

	t.Run("errors if extension doesn't exist on server", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 404,
    "message": "Extension 'com.dynatrace.extension.unknown' not found."
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

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetEnvironmentConfiguration(t.Context(), "com.dynatrace.extension.unknown")

		assert.Empty(t, resp)
		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.GetEnvironmentConfiguration(t.Context(), "com.dynatrace.extension.foo")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.ClientError{})
	})
}

func TestGetMonitoringConfiguration(t *testing.T) {
	t.Run("successfully returns monitoring configuration for requested IDs", func(t *testing.T) {
		getResponse := `{"objectId": "config-id-1", "value": {"enabled": true}}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/extensions/v2/extensions/com.dynatrace.extension.foo/monitoring-configurations/config-id-1", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: getResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", "config-id-1")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, getResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without extension name", func(t *testing.T) {
		client := extensions.NewClient(&rest.Client{})

		resp, err := client.GetMonitoringConfiguration(t.Context(), "", "config-id-1")

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "extensions", Field: "extension-name", Reason: "is empty"})
	})

	t.Run("errors if called without configuration ID", func(t *testing.T) {
		client := extensions.NewClient(&rest.Client{})

		resp, err := client.GetMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", "")

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "monitoring-configurations", Field: "configuration-id", Reason: "is empty"})
	})

	t.Run("errors if monitoring configuration with ID doesn't exist on server", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 404,
    "message": "Monitoring configuration 'false_ID' not found."
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

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", "false_ID")

		assert.Empty(t, resp)
		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.GetMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", "config-id-1")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.ClientError{})
	})
}

func TestCreateMonitoringConfiguration(t *testing.T) {
	given := `{"value": {"enabled": true, "description": "my config"}}`

	t.Run("successfully creates a monitoring configuration", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/extensions/v2/extensions/com.dynatrace.extension.foo/monitoring-configurations", req.URL.Path)
					requestBody, _ := io.ReadAll(req.Body)
					require.JSONEq(t, given, string(requestBody))
					return testutils.Response{ResponseCode: http.StatusCreated, ResponseBody: `{"objectId": "new-config-id"}`}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.CreateMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", []byte(given))

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("errors if called without extension name", func(t *testing.T) {
		client := extensions.NewClient(&rest.Client{})

		resp, err := client.CreateMonitoringConfiguration(t.Context(), "", []byte(given))

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "extensions", Field: "extension-name", Reason: "is empty"})
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

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.CreateMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", []byte(given))

		assert.Empty(t, resp)
		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.CreateMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", []byte(given))

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.ClientError{})
	})
}

func TestUpdateMonitoringConfiguration(t *testing.T) {
	given := `{"value": {"enabled": false, "description": "updated config"}}`

	t.Run("successfully updates a monitoring configuration", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/extensions/v2/extensions/com.dynatrace.extension.foo/monitoring-configurations/config-id-1", req.URL.Path)
					requestBody, _ := io.ReadAll(req.Body)
					require.JSONEq(t, given, string(requestBody))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: `{"objectId": "config-id-1"}`}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.UpdateMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", "config-id-1", []byte(given))

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without extension name", func(t *testing.T) {
		client := extensions.NewClient(&rest.Client{})

		resp, err := client.UpdateMonitoringConfiguration(t.Context(), "", "config-id-1", []byte(given))

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "extensions", Field: "extension-name", Reason: "is empty"})
	})

	t.Run("errors if called without configuration ID", func(t *testing.T) {
		client := extensions.NewClient(&rest.Client{})

		resp, err := client.UpdateMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", "", []byte(given))

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "monitoring-configurations", Field: "configuration-id", Reason: "is empty"})
	})

	t.Run("errors if server returns an error", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 404,
    "message": "Monitoring configuration 'false_ID' not found."
  }
}`

		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.UpdateMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", "false_ID", []byte(given))

		assert.Empty(t, resp)
		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.UpdateMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", "config-id-1", []byte(given))

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.ClientError{})
	})
}

func TestDeleteMonitoringConfiguration(t *testing.T) {
	t.Run("successfully deletes a monitoring configuration", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/extensions/v2/extensions/com.dynatrace.extension.foo/monitoring-configurations/config-id-1", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusNoContent}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		err := client.DeleteMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", "config-id-1")

		assert.NoError(t, err)
	})

	t.Run("errors if called without extension name", func(t *testing.T) {
		client := extensions.NewClient(&rest.Client{})

		err := client.DeleteMonitoringConfiguration(t.Context(), "", "config-id-1")

		assert.ErrorIs(t, err, api.ValidationError{Resource: "extensions", Field: "extension-name", Reason: "is empty"})
	})

	t.Run("errors if called without configuration ID", func(t *testing.T) {
		client := extensions.NewClient(&rest.Client{})

		err := client.DeleteMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", "")

		assert.ErrorIs(t, err, api.ValidationError{Resource: "monitoring-configurations", Field: "configuration-id", Reason: "is empty"})
	})

	t.Run("errors if monitoring configuration with ID doesn't exist on server", func(t *testing.T) {
		errorResponse := `{
  "error": {
    "code": 404,
    "message": "Monitoring configuration 'false_ID' not found."
  }
}`

		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.Client()))

		err := client.DeleteMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", "false_ID")

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := extensions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		err := client.DeleteMonitoringConfiguration(t.Context(), "com.dynatrace.extension.foo", "config-id-1")

		assert.ErrorAs(t, err, &api.ClientError{})
	})
}
