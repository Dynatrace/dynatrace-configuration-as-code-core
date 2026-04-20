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
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/segments"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	responses := []testutils.ResponseDef{
		{
			GET: func(t *testing.T, req *http.Request) testutils.Response {
				require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments:lean", req.URL.Path)
				return testutils.Response{
					ResponseCode: http.StatusOK,
					ResponseBody: apiResponse,
				}
			},
		},
	}
	server := testutils.NewHTTPTestServer(t, responses)
	defer server.Close()

	client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

	resp, err := client.List(t.Context())
	require.NoError(t, err)
	require.JSONEq(t, expected, string(resp.Data))
}

func TestGet(t *testing.T) {
	t.Run("when called without id parameter, returns a validation error", func(t *testing.T) {
		client := segments.NewClient(&rest.Client{})

		actual, err := client.Get(t.Context(), "")

		assert.Error(t, err)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "segments", Field: "id", Reason: "is empty"})
		assert.Empty(t, actual)
	})

	t.Run("ID doesn't exists on server returns API error", func(t *testing.T) {
		apiResponse := `{
  "error": {
    "code": 404,
    "message": "Segment not found",
    "errorDetails": []
  }
}`
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/some-id", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
						ResponseBody: apiResponse,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "some-id")

		assert.Empty(t, resp)

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
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/uid", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: apiResponse,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "uid")

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, apiResponse, string(resp.Data))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestGetAll(t *testing.T) {
	t.Run("getting list fails with error", func(t *testing.T) {
		apiResponse := `{ "err" : "something went wrong" }`
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments:lean", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusBadRequest,
						ResponseBody: apiResponse,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAll(t.Context())

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "segments", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, apiResponse, string(apiErr.Body))
	})

	t.Run("getting individual object from server fails and return error", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments:lean", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: `{"filterSegments": [{"uid": "pC7j2sEDzAQ"}]}`,
					}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/pC7j2sEDzAQ", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusInternalServerError,
						ResponseBody: `{ "err" : "something went wrong" }`,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAll(t.Context())

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "segments", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("successfully returned all configuration from server", func(t *testing.T) {
		listResponse := `{
  "filterSegments": [
    {"uid": "qW5qn449RsG"},
    {"uid": "pC7j2sEDzAQ"}
  ]
}
`
		getResponse1 := `{
      "uid": "qW5qn449RsG",
      "name": "dev_environment",
      "description": "only includes data of the dev environment",
      "variables": {"type": "query", "value": "fetch logs | limit 1"},
      "isPublic": false,
      "owner": "2f321c04-566e-4779-b576-3c033b8cd9e9",
      "allowedOperations": ["READ", "WRITE", "DELETE"],
      "version": 2
    }`
		getResponse2 := `   {
      "uid": "pC7j2sEDzAQ",
      "name": "dev_environment",
      "description": "only includes data of the dev environment",
      "variables": {"type": "query", "value": "fetch logs | limit 1"},
      "isPublic": false,
      "owner": "2f321c04-566e-4779-b576-3c033b8cd9e9",
      "allowedOperations": ["READ", "WRITE", "DELETE"],
      "version": 1
    }`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments:lean", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: listResponse,
					}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/qW5qn449RsG", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getResponse1,
					}
				},
			},
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/pC7j2sEDzAQ", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: getResponse2,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.GetAll(t.Context())

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, getResponse1, string(resp[0].Data))
		assert.Equal(t, getResponse2, string(resp[1].Data))
	})
}

func TestCreate(t *testing.T) {
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
	t.Run("error returned from response, expected error", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusBadGateway,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		_, err := client.Create(t.Context(), []byte(payload))

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "segments", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
	})
	t.Run("error returned from client, expected error", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		_, err := client.Create(t.Context(), []byte(payload))

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "segments", clientErr.Resource)
	})
	t.Run("successfully created new entity on server", func(t *testing.T) {
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
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusCreated,
						ResponseBody: apiResponse,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Create(t.Context(), []byte(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, apiResponse, string(resp.Data))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

func TestUpdate(t *testing.T) {
	t.Run("error returned from client, expected error", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		_, err := client.Update(t.Context(), "id", []byte(``))

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "segments", clientErr.Resource)
	})
	t.Run("id not provided, expecting validation error", func(t *testing.T) {
		client := segments.NewClient(&rest.Client{})

		_, err := client.Update(t.Context(), "", []byte(``))
		assert.ErrorIs(t, err, api.ValidationError{Resource: "segments", Field: "id", Reason: "is empty"})
	})
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

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/uid", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusBadGateway,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "uid", []byte(payload))

		require.Error(t, err)
		require.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "segments", clientErr.Resource)
		assert.Equal(t, "uid", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
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

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/"+uid, req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: apiExistingResource,
					}
				},
			},
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/"+uid, req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusNoContent,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), uid, []byte(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Empty(t, string(resp.Data))
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
	t.Run("successfully updated existing entity on server, payload without owner and uid", func(t *testing.T) {
		uid := "D82a1jdA23a"
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

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/"+uid, req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: apiExistingResource,
					}
				},
			},
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/"+uid, req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusNoContent,
					}
				},
				ValidateRequest: func(t *testing.T, req *http.Request) {
					assertRequestPayload(t, req, uid, "2f321c04-566e-4779-b576-3c033b8cd9e9")
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), uid, []byte(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Empty(t, string(resp.Data))
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
	t.Run("error test case, return a get response without owner", func(t *testing.T) {
		uid := "D82a1jdA23a"
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

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/"+uid, req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: apiExistingResource,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), uid, []byte(payload))

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "segments", Field: "owner", Reason: "is empty"})
	})

	t.Run("error test case, return a get response with an invalid version", func(t *testing.T) {
		uid := "D82a1jdA23a"
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
			  "version": 0
			}`

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/"+uid, req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: apiExistingResource,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), uid, []byte(payload))

		assert.Empty(t, resp)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "segments", Field: "version", Reason: "is invalid"})
	})

	t.Run("error test case, malformed payload provided to update", func(t *testing.T) {
		uid := "D82a1jdA23a"
		payload := `{///---....}`

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

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/"+uid, req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: apiExistingResource,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), uid, []byte(payload))

		assert.Empty(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "segments", runtimeErr.Resource)
		assert.Equal(t, uid, runtimeErr.Identifier)
		assert.Equal(t, "failed to add owner and UID", runtimeErr.Reason)
	})
}

func TestDelete(t *testing.T) {
	t.Run("error returned from client, expected error", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		_, err := client.Delete(t.Context(), "id")

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "segments", clientErr.Resource)
		assert.Equal(t, "id", clientErr.Identifier)
	})
	t.Run("error empty id provided, expected error", func(t *testing.T) {
		client := segments.NewClient(&rest.Client{})

		_, err := client.Delete(t.Context(), "")
		assert.ErrorIs(t, err, api.ValidationError{Resource: "segments", Field: "id", Reason: "is empty"})
	})
	t.Run("ID doesn't exists on server returns error", func(t *testing.T) {
		apiResponse := `{
	 "error": {
	   "code": 404,
	   "message": "Segment not found",
	   "errorDetails": []
	 }
	}`
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/uid", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
						ResponseBody: apiResponse,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Delete(t.Context(), "uid")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "segments", clientErr.Resource)
		assert.Equal(t, "uid", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, apiResponse, string(apiErr.Body))
	})

	t.Run("successfully deleted entity with ID from server", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/uid", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusNoContent,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := segments.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Delete(t.Context(), "uid")

		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusNoContent)
	})
}

func assertRequestPayload(t *testing.T, r *http.Request, expectedUID string, expectedOwner string) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		t.Error("invalid payload type")
	}
	var testRequest struct {
		Owner string `json:"owner"`
		UID   string `json:"uid"`
	}
	err = json.Unmarshal(data, &testRequest)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, testRequest.Owner, expectedOwner)
	assert.Equal(t, testRequest.UID, expectedUID)
}
