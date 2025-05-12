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
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/settings/permissions"
)

func TestNewClient(t *testing.T) {
	actual := permissions.NewClient(&rest.Client{})
	require.IsType(t, permissions.Client{}, *actual)
}

func TestClient_GetAllAccessors(t *testing.T) {
	t.Run("when called without id parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.GetAllAccessors(t.Context(), "")
		assert.Error(t, err)

		var errPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errPermissions)
		assert.ErrorAs(t, err, &permissions.ErrorMissingObjectID)
		assert.Equal(t, permissions.GET, errPermissions.Operation)
		assert.Empty(t, actual)
	})

	t.Run("successfully requesting all permissions for requested objectID", func(t *testing.T) {
		getResponse := `{
					"accessors": [
						{
						  "permissions": [
							"r"
						  ],
						  "accessor": {
							"type": "all-users"
						  }
						},
						{
						  "permissions": [
							"r",
							"w"
						  ],
						  "accessor": {
							"type": "group",
							"id": "4c75c5cb-4f85-4a49-811a-cdf9ae55fd4e"
						  }
						},
						{
						  "permissions": [
							"r",
							"w"
						  ],
						  "accessor": {
							"type": "user",
							"id": "b3d80429-98b7-44d7-b7ab-3ea453d2f18e"
						  }
						}
					  ]
					}`
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			t.Log(request.URL.String())
			require.Equal(t, http.MethodGet, request.Method)
			require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions", request.URL.Path)

			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte(getResponse))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.GetAllAccessors(t.Context(), "my-object-id")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, getResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("If settings object with ID doesn't exists on server returns error", func(t *testing.T) {
		errorResponse := `{"error": {"code": 404,"message": "Settings not found"}}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodGet:
				require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-object-id/permissions", request.URL.Path)

				writer.WriteHeader(http.StatusNotFound)
				writer.Write([]byte(errorResponse))
			default:
				require.Failf(t, "unexpected http call", "unexpected http call: %s %s", request.Method, request.URL)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.GetAllAccessors(t.Context(), "some-object-id")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("when connection to server fails, error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		server.Close()
		resp, err := client.GetAllAccessors(t.Context(), "some-object-id")
		assert.Empty(t, resp)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.Equal(t, permissions.GET, errorPermissions.Operation)
	})
}

func TestClient_GetAllUsersAccessor(t *testing.T) {
	t.Run("when called without id parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.GetAllUsersAccessor(t.Context(), "")
		assert.Empty(t, actual)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.ErrorIs(t, errorPermissions, permissions.ErrorMissingObjectID)
		assert.Equal(t, permissions.GET, errorPermissions.Operation)
	})

	t.Run("successfully requesting all-user permissions for requested objectID", func(t *testing.T) {
		getResponse := `{"permissions": ["r","w"],"accessor": {"type": "all-users"}}`
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			t.Log(request.URL.String())
			require.Equal(t, http.MethodGet, request.Method)
			require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions/all-users", request.URL.Path)

			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte(getResponse))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.GetAllUsersAccessor(t.Context(), "my-object-id")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, getResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("If settings object with ID doesn't exists on server returns error", func(t *testing.T) {
		errorResponse := `{"error": {"code": 404,"message": "Settings not found"}}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodGet:
				require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-object-id/permissions/all-users", request.URL.Path)

				writer.WriteHeader(http.StatusNotFound)
				writer.Write([]byte(errorResponse))
			default:
				require.Failf(t, "unexpected http call", "unexpected http call: %s %s", request.Method, request.URL)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.GetAllUsersAccessor(t.Context(), "some-object-id")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("when connection to server fails, error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		server.Close()
		resp, err := client.GetAllUsersAccessor(t.Context(), "some-object-id")
		assert.Empty(t, resp)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.Equal(t, permissions.GET, errorPermissions.Operation)
	})
}

func TestClient_GetAccessor(t *testing.T) {
	t.Run("when called without object id parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		resp, err := client.GetAccessor(t.Context(), "", "user", "user-id")
		assert.Empty(t, resp)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.ErrorIs(t, errorPermissions, permissions.ErrorMissingObjectID)
		assert.Equal(t, permissions.GET, errorPermissions.Operation)
	})

	t.Run("when called without accessor type parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.GetAccessor(t.Context(), "my-object-id", "", "user-id")
		assert.Empty(t, actual)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.ErrorIs(t, errorPermissions, permissions.ErrorMissingAccessorType)
		assert.Equal(t, permissions.GET, errorPermissions.Operation)
	})

	t.Run("when called without accessor id parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.GetAccessor(t.Context(), "my-object-id", "group", "")
		assert.Empty(t, actual)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.ErrorIs(t, errorPermissions, permissions.ErrorMissingAccessorID)
		assert.Equal(t, permissions.GET, errorPermissions.Operation)
	})

	t.Run("successfully requesting group permissions for requested objectID and groupID", func(t *testing.T) {
		getResponse := `{"permissions": ["r"],"accessor": {"type": "group","id": "4c75c5cb-4f85-4a49-811a-cdf9ae55fd4e"}},`
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			t.Log(request.URL.String())
			require.Equal(t, http.MethodGet, request.Method)
			require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions/group/4c75c5cb-4f85-4a49-811a-cdf9ae55fd4e", request.URL.Path)

			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte(getResponse))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.GetAccessor(t.Context(), "my-object-id", "group", "4c75c5cb-4f85-4a49-811a-cdf9ae55fd4e")

		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, getResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("If settings object with ID doesn't exists on server returns error", func(t *testing.T) {
		errorResponse := `{"error": {"code": 404,"message": "Settings not found"}}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodGet:
				require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-object-id/permissions/user/uid", request.URL.Path)

				writer.WriteHeader(http.StatusNotFound)
				writer.Write([]byte(errorResponse))
			default:
				require.Failf(t, "unexpected http call", "unexpected http call: %s %s", request.Method, request.URL)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.GetAccessor(t.Context(), "some-object-id", "user", "uid")
		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})

	t.Run("when connection to server fails, error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		server.Close()
		resp, err := client.GetAccessor(t.Context(), "some-object-id", "user", "uid")
		assert.Empty(t, resp)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.Equal(t, permissions.GET, errorPermissions.Operation)
	})
}

func TestClient_Create(t *testing.T) {
	t.Run("when called without id parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		resp, err := client.Create(t.Context(), "", nil)
		assert.Empty(t, resp)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.ErrorIs(t, errorPermissions, permissions.ErrorMissingObjectID)
		assert.Equal(t, permissions.POST, errorPermissions.Operation)
	})

	t.Run("successful creation of settings object permission with given payload", func(t *testing.T) {
		given := `{"accessor": {"id": "03c7e839-ee7e-4023-b5db-6da0dc9697bc","type": "user"},"permissions": ["r"]}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			t.Log(request.URL.String())
			require.Equal(t, http.MethodPost, request.Method)
			require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions", request.URL.Path)
			requestBody, _ := io.ReadAll(request.Body)
			require.JSONEq(t, given, string(requestBody))

			writer.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		actual, err := client.Create(t.Context(), "my-object-id", json.RawMessage(given))

		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
		assert.Equal(t, http.StatusCreated, actual.StatusCode)
	})

	t.Run("when connection to server fails, error is returned", func(t *testing.T) {
		given := `{"accessor": {"id": "03c7e839-ee7e-4023-b5db-6da0dc9697bc","type": "user"},"permissions": ["r"]}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		server.Close()
		resp, err := client.Create(t.Context(), "my-object-id", json.RawMessage(given))
		assert.Empty(t, resp)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.Equal(t, permissions.POST, errorPermissions.Operation)
	})
}

func TestClient_UpdateAllUsersAccessor(t *testing.T) {
	t.Run("when called without object id parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		resp, err := client.UpdateAllUsersAccessor(t.Context(), "", nil)
		assert.Empty(t, resp)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.ErrorIs(t, errorPermissions, permissions.ErrorMissingObjectID)
		assert.Equal(t, permissions.PUT, errorPermissions.Operation)
	})

	t.Run("successful permission update for settings object with given payload", func(t *testing.T) {
		given := `{"permissions": ["r"]}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodPut:
				require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions/all-users", request.URL.Path)
				requestBody, _ := io.ReadAll(request.Body)
				require.JSONEq(t, given, string(requestBody))

				writer.WriteHeader(http.StatusOK)
			default:
				require.Failf(t, "unexpected http call", "unexpected http call: %s %s", request.Method, request.URL)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.UpdateAllUsersAccessor(t.Context(), "my-object-id", json.RawMessage(given))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
	})

	t.Run("permission update for non existing settings object fails with error", func(t *testing.T) {
		get404Response := `{"error": {"code": 404,"message": "Settings not found"}}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodPut:
				require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-object-id/permissions/all-users", request.URL.Path)

				writer.WriteHeader(http.StatusNotFound)
				writer.Write([]byte(get404Response))
			default:
				require.Failf(t, "unexpected http call", "unexpected http call: %s %s", request.Method, request.URL)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.UpdateAllUsersAccessor(t.Context(), "some-object-id", nil)

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, get404Response, string(apiErr.Body))
	})

	t.Run("when connection to server fails, error is returned", func(t *testing.T) {
		given := `{"permissions": ["r"]}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		server.Close()
		resp, err := client.UpdateAllUsersAccessor(t.Context(), "my-object-id", json.RawMessage(given))
		assert.Empty(t, resp)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.Equal(t, permissions.PUT, errorPermissions.Operation)
	})
}

func TestClient_UpdateAccessor(t *testing.T) {
	t.Run("when called without object id parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		resp, err := client.UpdateAccessor(t.Context(), "", "user", "uid", nil)
		assert.Empty(t, resp)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.ErrorIs(t, errorPermissions, permissions.ErrorMissingObjectID)
		assert.Equal(t, permissions.PUT, errorPermissions.Operation)
	})

	t.Run("when called without accessor type parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		resp, err := client.UpdateAccessor(t.Context(), "object-id", "", "uid", nil)
		assert.Empty(t, resp)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.ErrorIs(t, errorPermissions, permissions.ErrorMissingAccessorType)
		assert.Equal(t, permissions.PUT, errorPermissions.Operation)
	})

	t.Run("when called without accessor id parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		resp, err := client.UpdateAccessor(t.Context(), "object-id", "user", "", nil)
		assert.Empty(t, resp)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.ErrorIs(t, errorPermissions, permissions.ErrorMissingAccessorID)
		assert.Equal(t, permissions.PUT, errorPermissions.Operation)
	})

	t.Run("successful permission update for settings object with given payload", func(t *testing.T) {
		given := `{"permissions": ["r"]}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodPut:
				require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions/user/03c7e839-ee7e-4023-b5db-6da0dc9697bc", request.URL.Path)
				requestBody, _ := io.ReadAll(request.Body)
				require.JSONEq(t, given, string(requestBody))

				writer.WriteHeader(http.StatusOK)
			default:
				require.Failf(t, "unexpected http call", "unexpected http call: %s %s", request.Method, request.URL)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.UpdateAccessor(t.Context(), "my-object-id", "user", "03c7e839-ee7e-4023-b5db-6da0dc9697bc", json.RawMessage(given))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)

	})

	t.Run("permission update for non existing settings object fails with error", func(t *testing.T) {
		get404Response := `{"error": {"code": 404,"message": "Settings not found"}}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodPut:
				require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-object-id/permissions/user/uid", request.URL.Path)

				writer.WriteHeader(http.StatusNotFound)
				writer.Write([]byte(get404Response))
			default:
				require.Failf(t, "unexpected http call", "unexpected http call: %s %s", request.Method, request.URL)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.UpdateAccessor(t.Context(), "some-object-id", "user", "uid", nil)

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, get404Response, string(apiErr.Body))
	})

	t.Run("when connection to server fails, error is returned", func(t *testing.T) {
		given := `{"permissions": ["r"]}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
		url, _ := url.Parse(server.URL + "/invalid-path")
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		server.Close()
		resp, err := client.UpdateAccessor(t.Context(), "some-object-id", "user", "uid", json.RawMessage(given))
		assert.Empty(t, resp)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.Equal(t, permissions.PUT, errorPermissions.Operation)
	})
}

func TestClient_DeleteAccessor(t *testing.T) {
	t.Run("when called without object id parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.DeleteAccessor(t.Context(), "", "user", "uid")
		assert.Empty(t, actual)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.ErrorIs(t, errorPermissions, permissions.ErrorMissingObjectID)
		assert.Equal(t, permissions.DELETE, errorPermissions.Operation)
	})

	t.Run("when called without accessor type parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.DeleteAccessor(t.Context(), "object-id", "", "uid")
		assert.Empty(t, actual)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.ErrorIs(t, errorPermissions, permissions.ErrorMissingAccessorType)
		assert.Equal(t, permissions.DELETE, errorPermissions.Operation)
	})

	t.Run("when called without accessor id parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.DeleteAccessor(t.Context(), "object-id", "user", "")
		assert.Empty(t, actual)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.ErrorIs(t, errorPermissions, permissions.ErrorMissingAccessorID)
		assert.Equal(t, permissions.DELETE, errorPermissions.Operation)
	})

	t.Run("successfully deleted permissions for settings object", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodDelete:
				require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions/user/uid", request.URL.Path)

				writer.WriteHeader(http.StatusOK)
			default:
				require.Failf(t, "unexpected http call", "unexpected http call: %s %s", request.Method, request.URL)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		actual, err := client.DeleteAccessor(t.Context(), "my-object-id", "user", "uid")

		assert.NoError(t, err)
		assert.Equal(t, actual.StatusCode, http.StatusOK)
	})

	t.Run("If settings object with ID doesn't exists on server returns an error", func(t *testing.T) {
		get404Response := `{"error": {"code": 404,"message": "Settings not found"}}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodDelete:
				require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-unknown-id/permissions/user/uid", request.URL.Path)

				writer.WriteHeader(http.StatusNotFound)
				writer.Write([]byte(get404Response))
			default:
				require.Failf(t, "unexpected http call", "unexpected http call: %s %s", request.Method, request.URL)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		actual, err := client.DeleteAccessor(t.Context(), "some-unknown-id", "user", "uid")

		assert.Empty(t, actual)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		errors.As(err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, get404Response, string(apiErr.Body))
	})

	t.Run("when connection to server fails, error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
		url, _ := url.Parse(server.URL + "/invalid-path")
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		server.Close()
		actual, err := client.DeleteAccessor(t.Context(), "some-unknown-id", "user", "uid")
		assert.Empty(t, actual)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.Equal(t, permissions.DELETE, errorPermissions.Operation)
	})
}

func TestClient_DeleteAllUsersAccessor(t *testing.T) {
	t.Run("when called without object id parameter, returns an error", func(t *testing.T) {
		client := permissions.NewClient(&rest.Client{})

		actual, err := client.DeleteAllUsersAccessor(t.Context(), "")
		assert.Empty(t, actual)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.ErrorIs(t, errorPermissions, permissions.ErrorMissingObjectID)
		assert.Equal(t, permissions.DELETE, errorPermissions.Operation)
	})

	t.Run("successfully deleted permissions for settings object", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodDelete:
				require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/my-object-id/permissions/all-users", request.URL.Path)

				writer.WriteHeader(http.StatusOK)
			default:
				require.Failf(t, "unexpected http call", "unexpected http call: %s %s", request.Method, request.URL)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		actual, err := client.DeleteAllUsersAccessor(t.Context(), "my-object-id")

		assert.NoError(t, err)
		assert.Equal(t, actual.StatusCode, http.StatusOK)
	})

	t.Run("If settings object with ID doesn't exists on server returns an error", func(t *testing.T) {
		get404Response := `{"error": {"code": 404,"message": "Settings not found"}}`

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodDelete:
				require.Equal(t, "/platform/classic/environment-api/v2/settings/objects/some-unknown-id/permissions/all-users", request.URL.Path)

				writer.WriteHeader(http.StatusNotFound)
				writer.Write([]byte(get404Response))
			default:
				require.Failf(t, "unexpected http call", "unexpected http call: %s %s", request.Method, request.URL)
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		actual, err := client.DeleteAllUsersAccessor(t.Context(), "some-unknown-id")

		assert.Empty(t, actual)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		errors.As(err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, get404Response, string(apiErr.Body))
	})

	t.Run("when connection to server fails, error is returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
		url, _ := url.Parse(server.URL + "/invalid-path")
		client := permissions.NewClient(rest.NewClient(url, server.Client()))

		server.Close()
		actual, err := client.DeleteAllUsersAccessor(t.Context(), "some-unknown-id")
		assert.Empty(t, actual)
		assert.Error(t, err)

		var errorPermissions permissions.ErrorPermissions
		assert.ErrorAs(t, err, &errorPermissions)
		assert.Equal(t, permissions.DELETE, errorPermissions.Operation)
	})
}
