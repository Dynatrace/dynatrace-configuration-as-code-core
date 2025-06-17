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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/segments"
	intErr "github.com/dynatrace/dynatrace-configuration-as-code-core/internal/errors"
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments:lean", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(apiResponse))
	}))
	defer server.Close()

	url, _ := url.Parse(server.URL)
	client := segments.NewClient(rest.NewClient(url, server.Client()))

	resp, err := client.List(t.Context())
	require.NoError(t, err)
	require.JSONEq(t, expected, string(resp.Data))
}

func TestGet(t *testing.T) {
	t.Run("when called without id parameter, returns an validation error", func(t *testing.T) {
		client := segments.NewClient(&rest.Client{})

		actual, err := client.Get(t.Context(), "")

		assert.Error(t, err)
		assert.ErrorIs(t, err, intErr.ErrorValidation{Field: "id", Reason: "is empty"})
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
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/some-id", r.URL.Path)

			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(apiResponse))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

		id := "some-id"
		resp, err := client.Get(t.Context(), id)

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
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/uid", r.URL.Path)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(apiResponse))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

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
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments:lean", r.URL.Path)

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(apiResponse))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.GetAll(t.Context())

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		errors.As(err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, apiResponse, string(apiErr.Body))
	})

	t.Run("getting individual object from server fails and return error", func(t *testing.T) {
		response := map[int]string{
			0: `{"filterSegments": [{"uid": "pC7j2sEDzAQ"}]}`,
			1: `{ "err" : "something went wrong" }`,
		}

		i := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments:lean", r.URL.Path)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(response[i]))
			i++
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.GetAll(t.Context())

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		errors.As(err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	})

	t.Run("successfully returned all configuration from server", func(t *testing.T) {
		expectedEnding := map[int]string{
			0: `:lean`,
			1: `/qW5qn449RsG`,
			2: `/pC7j2sEDzAQ`,
		}
		expectedResponse := map[int]string{
			0: `{
  "filterSegments": [
    {"uid": "qW5qn449RsG"},
    {"uid": "pC7j2sEDzAQ"}
  ]
}
`,
			1: `{
      "uid": "qW5qn449RsG",
      "name": "dev_environment",
      "description": "only includes data of the dev environment",
      "variables": {"type": "query", "value": "fetch logs | limit 1"},
      "isPublic": false,
      "owner": "2f321c04-566e-4779-b576-3c033b8cd9e9",
      "allowedOperations": ["READ", "WRITE", "DELETE"],
      "version": 2
    }`,
			2: `   {
      "uid": "pC7j2sEDzAQ",
      "name": "dev_environment",
      "description": "only includes data of the dev environment",
      "variables": {"type": "query", "value": "fetch logs | limit 1"},
      "isPublic": false,
      "owner": "2f321c04-566e-4779-b576-3c033b8cd9e9",
      "allowedOperations": ["READ", "WRITE", "DELETE"],
      "version": 1
    }`,
		}
		i := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			url := "/platform/storage/filter-segments/v1/filter-segments" + expectedEnding[i]
			require.Equal(t, url, r.URL.Path)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(expectedResponse[i]))
			i++
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.GetAll(t.Context())

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		for k := 1; k < len(expectedResponse); k++ {
			assert.Equal(t, expectedResponse[k], string(resp[k-1].Data))
		}
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
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

		_, err := client.Create(t.Context(), []byte(payload))
		assert.Error(t, err)
	})
	t.Run("error returned from client, expected error", func(t *testing.T) {
		httpClient := &http.Client{}
		path := &url.URL{}
		client := segments.NewClient(rest.NewClient(path, httpClient))

		_, err := client.Create(t.Context(), []byte(payload))
		assert.Error(t, err)
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
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments", r.URL.Path)

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(apiResponse))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.Create(t.Context(), []byte(payload))

		assert.NotEmpty(t, resp)
		assert.NoError(t, err)
		assert.Equal(t, apiResponse, string(resp.Data))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}

func TestUpdate(t *testing.T) {
	t.Run("error returned from client, expected error", func(t *testing.T) {
		httpClient := &http.Client{}
		path := &url.URL{}
		client := segments.NewClient(rest.NewClient(path, httpClient))

		_, err := client.Update(t.Context(), "id", []byte(``))
		assert.Error(t, err)
	})
	t.Run("id not provided, expecting validation error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("should failt at id validation")
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

		_, err := client.Update(t.Context(), "", []byte(``))
		assert.Error(t, err)
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

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/uid", r.URL.Path)

			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(``))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.Update(t.Context(), "uid", []byte(payload))

		require.Error(t, err)
		require.Empty(t, resp)
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

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/"+uid, r.URL.Path)
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(apiExistingResource))
				break
			case http.MethodPut:
				w.WriteHeader(http.StatusNoContent)
				break
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

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

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/"+uid, r.URL.Path)
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(apiExistingResource))
				break
			case http.MethodPut:
				assertRequestPayload(t, r, uid, "2f321c04-566e-4779-b576-3c033b8cd9e9")
				w.WriteHeader(http.StatusNoContent)
				break
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

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

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/"+uid, r.URL.Path)
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(apiExistingResource))
				break
			case http.MethodPut:
				t.Errorf("should failt at owner validation")
				break
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.Update(t.Context(), uid, []byte(payload))

		assert.Empty(t, resp)
		assert.Error(t, err)
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

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/"+uid, r.URL.Path)
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(apiExistingResource))
				break
			case http.MethodPut:
				t.Errorf("should failt at unmarshall payload")
				break
			}
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.Update(t.Context(), uid, []byte(payload))

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestDelete(t *testing.T) {
	t.Run("error returned from client, expected error", func(t *testing.T) {
		httpClient := &http.Client{}
		path := &url.URL{}
		client := segments.NewClient(rest.NewClient(path, httpClient))

		_, err := client.Delete(t.Context(), "id")
		assert.Error(t, err)
	})
	t.Run("error empty id provided, expected error", func(t *testing.T) {
		httpClient := &http.Client{}
		path := &url.URL{}
		client := segments.NewClient(rest.NewClient(path, httpClient))

		_, err := client.Delete(t.Context(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id")
	})
	t.Run("ID doesn't exists on server returns error", func(t *testing.T) {
		apiResponse := `{
	 "error": {
	   "code": 404,
	   "message": "Segment not found",
	   "errorDetails": []
	 }
	}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodDelete, r.Method)
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/uid", r.URL.Path)

			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(apiResponse))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

		resp, err := client.Delete(t.Context(), "uid")

		assert.Empty(t, resp)
		assert.ErrorAs(t, err, &api.APIError{})

		var apiErr api.APIError
		errors.As(err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.Equal(t, apiResponse, string(apiErr.Body))
	})

	t.Run("successfully deleted entity with ID from server", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodDelete, r.Method)
			require.Equal(t, "/platform/storage/filter-segments/v1/filter-segments/uid", r.URL.Path)

			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(``))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		client := segments.NewClient(rest.NewClient(url, server.Client()))

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
