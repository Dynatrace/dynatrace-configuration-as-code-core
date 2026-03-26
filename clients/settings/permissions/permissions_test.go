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

package permissions_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/settings/permissions"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

func TestNewClient(t *testing.T) {
	actual := permissions.NewClient(&rest.Client{})
	require.IsType(t, permissions.Client{}, *actual)
}

func TestGetAllAccessors(t *testing.T) {
	t.Run("successfully returns all permissions for requested objectID", func(t *testing.T) {
		getResponse := `{
			"accessors": [
				{"permissions": ["r"], "accessor": {"type": "all-users"}},
				{"permissions": ["r","w"], "accessor": {"type": "group", "id": "4c75c5cb-4f85-4a49-811a-cdf9ae55fd4e"}},
				{"permissions": ["r","w"], "accessor": {"type": "user", "id": "b3d80429-98b7-44d7-b7ab-3ea453d2f18e"}}
			]
		}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions", req.URL.Path)
					require.Equal(t, "true", req.URL.Query().Get("adminAccess"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: getResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAllAccessors(t.Context(), "my-object-id", true)

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without objectID parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.GetAllAccessors(t.Context(), "", true)

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "objectID", Reason: "is empty"})
	})

	t.Run("errors if settings object with ID doesn't exist on server", func(t *testing.T) {
		errorResponse := `{"error": {"code": 404,"message": "Settings not found"}}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-object-id/permissions", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAllAccessors(t.Context(), "some-object-id", true)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "some-object-id", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.GetAllAccessors(t.Context(), "some-object-id", true)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "some-object-id", clientErr.Identifier)
	})

	t.Run("sends adminAccess=false when specified", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "false", req.URL.Query().Get("adminAccess"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: `{"accessors":[]}`}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAllAccessors(t.Context(), "my-object-id", false)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestGetAllUsersAccessor(t *testing.T) {
	t.Run("successfully returns all-users permissions for requested objectID", func(t *testing.T) {
		getResponse := `{"permissions": ["r","w"],"accessor": {"type": "all-users"}}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions/all-users", req.URL.Path)
					require.Equal(t, "true", req.URL.Query().Get("adminAccess"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: getResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAllUsersAccessor(t.Context(), "my-object-id", true)

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, getResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without objectID parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.GetAllUsersAccessor(t.Context(), "", true)

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "objectID", Reason: "is empty"})
	})

	t.Run("errors if settings object doesn't exist on server", func(t *testing.T) {
		errorResponse := `{"error": {"code": 404,"message": "Settings not found"}}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-object-id/permissions/all-users", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAllUsersAccessor(t.Context(), "some-object-id", true)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "some-object-id", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.GetAllUsersAccessor(t.Context(), "some-object-id", true)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "some-object-id", clientErr.Identifier)
	})
}

func TestGetAccessor(t *testing.T) {
	t.Run("successfully returns accessor permissions for requested objectID", func(t *testing.T) {
		getResponse := `{"permissions": ["r"],"accessor": {"type": "group","id": "4c75c5cb-4f85-4a49-811a-cdf9ae55fd4e"}}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions/group/4c75c5cb-4f85-4a49-811a-cdf9ae55fd4e", req.URL.Path)
					require.Equal(t, "true", req.URL.Query().Get("adminAccess"))
					return testutils.Response{ResponseCode: http.StatusOK, ResponseBody: getResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAccessor(t.Context(), "my-object-id", "group", "4c75c5cb-4f85-4a49-811a-cdf9ae55fd4e", true)

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, getResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if called without objectID parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		resp, err := client.GetAccessor(t.Context(), "", "user", "user-id", true)

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "objectID", Reason: "is empty"})
	})

	t.Run("errors if called without accessorType parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.GetAccessor(t.Context(), "my-object-id", "", "user-id", true)

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "accessorType", Reason: "is empty"})
	})

	t.Run("errors if called without accessorID parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.GetAccessor(t.Context(), "my-object-id", "group", "", true)

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "accessorID", Reason: "is empty"})
	})

	t.Run("errors if settings object doesn't exist on server", func(t *testing.T) {
		errorResponse := `{"error": {"code": 404,"message": "Settings not found"}}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-object-id/permissions/user/uid", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAccessor(t.Context(), "some-object-id", "user", "uid", true)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "some-object-id", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.GetAccessor(t.Context(), "some-object-id", "user", "uid", true)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "some-object-id", clientErr.Identifier)
	})
}

func TestCreate(t *testing.T) {
	t.Run("successfully creates a settings object permission", func(t *testing.T) {
		given := `{"accessor": {"id": "03c7e839-ee7e-4023-b5db-6da0dc9697bc","type": "user"},"permissions": ["r"]}`

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions", req.URL.Path)
					require.Equal(t, "true", req.URL.Query().Get("adminAccess"))
					requestBody, _ := io.ReadAll(req.Body)
					require.JSONEq(t, given, string(requestBody))
					return testutils.Response{ResponseCode: http.StatusCreated}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.Create(t.Context(), "my-object-id", true, []byte(given))

		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
		assert.Equal(t, http.StatusCreated, actual.StatusCode)
	})

	t.Run("errors if called without objectID parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		resp, err := client.Create(t.Context(), "", true, nil)

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "objectID", Reason: "is empty"})
	})

	t.Run("errors if server returns an error", func(t *testing.T) {
		errorResponse := `{"error": {"code": 400,"message": "Invalid request body"}}`

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, _ *http.Request) testutils.Response {
					return testutils.Response{ResponseCode: http.StatusBadRequest, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.Create(t.Context(), "my-object-id", true, []byte(`{}`))

		assert.Empty(t, actual)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "my-object-id", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Create(t.Context(), "my-object-id", true, []byte(`{}`))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "my-object-id", clientErr.Identifier)
	})
}

func TestUpdateAllUsersAccessor(t *testing.T) {
	t.Run("successfully updates all-users permission", func(t *testing.T) {
		given := `{"permissions": ["r"]}`

		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions/all-users", req.URL.Path)
					require.Equal(t, "true", req.URL.Query().Get("adminAccess"))
					requestBody, _ := io.ReadAll(req.Body)
					require.JSONEq(t, given, string(requestBody))
					return testutils.Response{ResponseCode: http.StatusOK}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.UpdateAllUsersAccessor(t.Context(), "my-object-id", true, []byte(given))

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
	})

	t.Run("errors if called without objectID parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		resp, err := client.UpdateAllUsersAccessor(t.Context(), "", true, nil)

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "objectID", Reason: "is empty"})
	})

	t.Run("errors if settings object doesn't exist on server", func(t *testing.T) {
		errorResponse := `{"error": {"code": 404,"message": "Settings not found"}}`

		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-object-id/permissions/all-users", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.UpdateAllUsersAccessor(t.Context(), "some-object-id", true, nil)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "some-object-id", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.UpdateAllUsersAccessor(t.Context(), "my-object-id", true, []byte(`{}`))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "my-object-id", clientErr.Identifier)
	})

	t.Run("sends adminAccess=false when specified", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "false", req.URL.Query().Get("adminAccess"))
					return testutils.Response{ResponseCode: http.StatusOK}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.UpdateAllUsersAccessor(t.Context(), "my-object-id", false, []byte(`{}`))

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
	})
}

func TestUpdateAccessor(t *testing.T) {
	t.Run("successfully updates accessor permission", func(t *testing.T) {
		given := `{"permissions": ["r"]}`

		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions/user/03c7e839-ee7e-4023-b5db-6da0dc9697bc", req.URL.Path)
					require.Equal(t, "true", req.URL.Query().Get("adminAccess"))
					requestBody, _ := io.ReadAll(req.Body)
					require.JSONEq(t, given, string(requestBody))
					return testutils.Response{ResponseCode: http.StatusOK}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.UpdateAccessor(t.Context(), "my-object-id", "user", "03c7e839-ee7e-4023-b5db-6da0dc9697bc", true, []byte(given))

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
	})

	t.Run("errors if called without objectID parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		resp, err := client.UpdateAccessor(t.Context(), "", "user", "uid", true, nil)

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "objectID", Reason: "is empty"})
	})

	t.Run("errors if called without accessorType parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		resp, err := client.UpdateAccessor(t.Context(), "object-id", "", "uid", true, nil)

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "accessorType", Reason: "is empty"})
	})

	t.Run("errors if called without accessorID parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		resp, err := client.UpdateAccessor(t.Context(), "object-id", "user", "", true, nil)

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "accessorID", Reason: "is empty"})
	})

	t.Run("errors if settings object doesn't exist on server", func(t *testing.T) {
		errorResponse := `{"error": {"code": 404,"message": "Settings not found"}}`

		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-object-id/permissions/user/uid", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.UpdateAccessor(t.Context(), "some-object-id", "user", "uid", true, nil)

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "some-object-id", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.UpdateAccessor(t.Context(), "some-object-id", "user", "uid", true, []byte(`{}`))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "some-object-id", clientErr.Identifier)
	})
}

func TestDeleteAccessor(t *testing.T) {
	t.Run("successfully deletes accessor permission", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions/user/uid", req.URL.Path)
					require.Equal(t, "true", req.URL.Query().Get("adminAccess"))
					return testutils.Response{ResponseCode: http.StatusOK}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.DeleteAccessor(t.Context(), "my-object-id", "user", "uid", true)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, actual.StatusCode)
	})

	t.Run("errors if called without objectID parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.DeleteAccessor(t.Context(), "", "user", "uid", true)

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "objectID", Reason: "is empty"})
	})

	t.Run("errors if called without accessorType parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.DeleteAccessor(t.Context(), "object-id", "", "uid", true)

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "accessorType", Reason: "is empty"})
	})

	t.Run("errors if called without accessorID parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.DeleteAccessor(t.Context(), "object-id", "user", "", true)

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "accessorID", Reason: "is empty"})
	})

	t.Run("errors if settings object doesn't exist on server", func(t *testing.T) {
		errorResponse := `{"error": {"code": 404,"message": "Settings not found"}}`

		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-unknown-id/permissions/user/uid", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.DeleteAccessor(t.Context(), "some-unknown-id", "user", "uid", true)

		assert.Empty(t, actual)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "some-unknown-id", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		actual, err := client.DeleteAccessor(t.Context(), "some-unknown-id", "user", "uid", true)

		assert.Empty(t, actual)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "some-unknown-id", clientErr.Identifier)
	})

	t.Run("sends adminAccess=false when specified", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "false", req.URL.Query().Get("adminAccess"))
					return testutils.Response{ResponseCode: http.StatusOK}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.DeleteAccessor(t.Context(), "my-object-id", "user", "uid", false)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, actual.StatusCode)
	})
}

func TestDeleteAllUsersAccessor(t *testing.T) {
	t.Run("successfully deletes all-users accessor permission", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions/all-users", req.URL.Path)
					require.Equal(t, "true", req.URL.Query().Get("adminAccess"))
					return testutils.Response{ResponseCode: http.StatusOK}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.DeleteAllUsersAccessor(t.Context(), "my-object-id", true)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, actual.StatusCode)
	})

	t.Run("errors if called without objectID parameter", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.DeleteAllUsersAccessor(t.Context(), "", true)

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "permissions", Field: "objectID", Reason: "is empty"})
	})

	t.Run("errors if settings object doesn't exist on server", func(t *testing.T) {
		errorResponse := `{"error": {"code": 404,"message": "Settings not found"}}`

		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-unknown-id/permissions/all-users", req.URL.Path)
					return testutils.Response{ResponseCode: http.StatusNotFound, ResponseBody: errorResponse}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.Client()))

		actual, err := client.DeleteAllUsersAccessor(t.Context(), "some-unknown-id", true)

		assert.Empty(t, actual)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "some-unknown-id", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := permissions.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		actual, err := client.DeleteAllUsersAccessor(t.Context(), "some-unknown-id", true)

		assert.Empty(t, actual)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "permissions", clientErr.Resource)
		assert.Equal(t, "some-unknown-id", clientErr.Identifier)
	})
}
