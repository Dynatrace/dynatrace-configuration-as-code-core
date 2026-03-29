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

package buckets_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/buckets"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
)

const (
	activeBucketResponse = `{
 "bucketName": "bucket name",
 "table": "metrics",
 "displayName": "Default metrics (15 months)",
 "status": "active",
 "retentionDays": 462,
 "metricInterval": "PT1M",
 "version": 1
}`

	creatingBucketResponse = `{
 "bucketName": "bucket name",
 "table": "metrics",
 "displayName": "Default metrics (15 months)",
 "status": "creating",
 "retentionDays": 462,
 "metricInterval": "PT1M",
 "version": 1
}`
	updatingBucketResponse = `{
 "bucketName": "bucket name",
 "table": "metrics",
 "displayName": "Default metrics (15 months)",
 "status": "updating",
 "retentionDays": 462,
 "metricInterval": "PT1M",
 "version": 1
}`

	deletingBucketResponse = `{
 "bucketName": "bucket name",
 "table": "metrics",
 "displayName": "Default metrics (15 months)",
 "status": "deleting",
 "retentionDays": 462,
 "metricInterval": "PT1M",
 "version": 1
}`
)

func TestNewClient(t *testing.T) {
	actual := buckets.NewClient(&rest.Client{})
	require.IsType(t, buckets.Client{}, *actual)
}

func TestGet(t *testing.T) {
	t.Run("successfully fetch a bucket", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/management/v1/bucket-definitions/bucket name", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: activeBucketResponse,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "bucket name")
		assert.NoError(t, err)
		assert.Equal(t, activeBucketResponse, string(resp.Data))
	})

	t.Run("errors if called without bucketName parameter", func(t *testing.T) {
		client := buckets.NewClient(&rest.Client{})

		actual, err := client.Get(t.Context(), "")

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "buckets", Field: "bucketName", Reason: "is empty"})
	})

	t.Run("errors if server returns non-2xx", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/management/v1/bucket-definitions/bucket name", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
						ResponseBody: "{}",
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Get(t.Context(), "bucket name")

		assert.Zero(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "buckets", clientErr.Resource)
		assert.Equal(t, "bucket name", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Get(t.Context(), "some-bucket")

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "buckets", clientErr.Resource)
		assert.Equal(t, "some-bucket", clientErr.Identifier)
	})
}

func TestList(t *testing.T) {
	t.Run("successfully fetch a list of buckets", func(t *testing.T) {
		const bucket1 = `{
		"bucketName": "bucket name",
		"table": "metrics",
		"displayName": "Default metrics (15 months)",
		"status": "active",
		"retentionDays": 462,
		"metricInterval": "PT1M",
		"version": 1
	}`
		const bucket2 = `{
		"bucketName": "another name",
		"table": "metrics",
		"displayName": "Some logs",
		"status": "active",
		"retentionDays": 31,
		"metricInterval": "PT2M",
		"version": 42
	}`
		payload := fmt.Sprintf(`{
		"buckets": [
			%s,
			%s
		]
	}`, bucket1, bucket2)

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/management/v1/bucket-definitions", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payload,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())

		assert.NoError(t, err)
		assert.Equal(t, resp[0].Data, []byte(payload))
		assert.ElementsMatch(t, resp.All(), [][]byte{[]byte(bucket1), []byte(bucket2)})
	})

	t.Run("successfully returns empty response if no buckets exist", func(t *testing.T) {
		const payload = `{ "buckets": [] }`
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/management/v1/bucket-definitions", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payload,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())

		assert.NoError(t, err)
		assert.Equal(t, resp[0].Data, []byte(payload))
		assert.Empty(t, resp.All())
	})

	t.Run("errors if server returns non-2xx", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/management/v1/bucket-definitions", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
						ResponseBody: "{}",
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())
		assert.Len(t, resp, 0)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "buckets", clientErr.Resource)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.List(t.Context())

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "buckets", clientErr.Resource)

		assert.Empty(t, resp)
	})

	t.Run("errors if response is invalid JSON", func(t *testing.T) {
		const payload = `{ buckets: [] }`
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payload,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "buckets", runtimeErr.Resource)
		assert.Equal(t, "unmarshalling failed", runtimeErr.Reason)

		assert.Empty(t, resp)
	})
}

func TestDelete(t *testing.T) {
	t.Run("errors if bucket name is empty", func(t *testing.T) {
		client := buckets.NewClient(&rest.Client{})

		actual, err := client.Delete(t.Context(), "")

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "buckets", Field: "bucketName", Reason: "is empty"})
	})

	t.Run("successfully deletes bucket", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/management/v1/bucket-definitions/bucket name", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusAccepted,
						ResponseBody: deletingBucketResponse,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Delete(t.Context(), "bucket name")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
		assert.Equal(t, deletingBucketResponse, string(resp.Data))
	})

	t.Run("errors if bucket to be deleted is not found", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/management/v1/bucket-definitions/bucket name", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Delete(t.Context(), "bucket name")

		assert.Zero(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "buckets", clientErr.Resource)
		assert.Equal(t, "bucket name", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Delete(t.Context(), "bucket name")

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodDelete, clientErr.Operation)
		assert.Equal(t, "buckets", clientErr.Resource)
		assert.Equal(t, "bucket name", clientErr.Identifier)

		assert.Zero(t, resp)
	})
}

func TestCreate(t *testing.T) {
	const someBucketData = `{
 "bucketName": "bucket name",
 "table": "metrics",
 "displayName": "Default metrics (15 months)",
 "retentionDays": 462,
 "metricInterval": "PT1M",
 "version": 1
}`

	t.Run("successfully creates bucket", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/management/v1/bucket-definitions", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusCreated,
						ResponseBody: creatingBucketResponse,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Create(t.Context(), "bucket name", []byte(someBucketData))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, creatingBucketResponse, string(resp.Data))
	})

	t.Run("errors if HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Create(t.Context(), "bucket name", []byte(someBucketData))

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "buckets", clientErr.Resource)
		assert.Equal(t, "bucket name", clientErr.Identifier)

		assert.Zero(t, resp)
	})

	t.Run("errors if data is invalid JSON", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Create(t.Context(), "bucket name", []byte("-)§/$/(="))

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "buckets", runtimeErr.Resource)
		assert.Equal(t, "bucket name", runtimeErr.Identifier)
		assert.Equal(t, "failed to set bucket name in payload", runtimeErr.Reason)

		assert.Zero(t, resp)
	})

	t.Run("errors if server returns non-2xx", func(t *testing.T) {
		errorResponse := `{"error":{"code":400,"message":"Invalid request body"}}`
		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusBadRequest,
						ResponseBody: errorResponse,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Create(t.Context(), "bucket name", []byte(someBucketData))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPost, clientErr.Operation)
		assert.Equal(t, "buckets", clientErr.Resource)
		assert.Equal(t, "bucket name", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, errorResponse, string(apiErr.Body))
	})
}

func TestUpdate(t *testing.T) {
	t.Run("errors if called without bucketName parameter", func(t *testing.T) {
		client := buckets.NewClient(&rest.Client{})

		actual, err := client.Update(t.Context(), "", nil)

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, api.ValidationError{Resource: "buckets", Field: "bucketName", Reason: "is empty"})
	})

	t.Run("errors if GET fails with server error", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/management/v1/bucket-definitions/bucket name", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusForbidden,
						ResponseBody: "expected error, we don't want to get",
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "bucket name", []byte("{}"))

		assert.Zero(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "buckets", clientErr.Resource)
		assert.Equal(t, "bucket name", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
	})

	t.Run("errors if GET HTTP request fails", func(t *testing.T) {
		server := testutils.NewHTTPTestServer(t, []testutils.ResponseDef{})
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		resp, err := client.Update(t.Context(), "bucket name", []byte("{}"))

		assert.Empty(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodGet, clientErr.Operation)
		assert.Equal(t, "buckets", clientErr.Resource)
		assert.Equal(t, "bucket name", clientErr.Identifier)
	})

	t.Run("successfully updates bucket", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/management/v1/bucket-definitions/bucket name", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: activeBucketResponse,
					}
				},
			},
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/management/v1/bucket-definitions/bucket name", req.URL.Path)
					require.Equal(t, "1", req.URL.Query().Get("optimistic-locking-version"))
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: updatingBucketResponse,
					}
				},
				ValidateRequest: func(t *testing.T, req *http.Request) {
					data, err := io.ReadAll(req.Body)
					require.NoError(t, err)

					m := map[string]any{}
					require.NoError(t, json.Unmarshal(data, &m))

					assert.Equal(t, "bucket name", m["bucketName"])
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "bucket name", []byte("{}"))
		assert.NoError(t, err)

		m := map[string]any{}
		err = json.Unmarshal(resp.Data, &m)
		assert.NoError(t, err)
		assert.Equal(t, "bucket name", m["bucketName"])
	})

	t.Run("skips update when bucket is unmodified", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					require.Equal(t, "/platform/storage/management/v1/bucket-definitions/bucket name", req.URL.Path)
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: activeBucketResponse,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))
		data := []byte(`{
	 "bucketName": "bucket name",
	 "table": "metrics",
	 "displayName": "Default metrics (15 months)",
	 "retentionDays": 462,
	 "metricInterval": "PT1M"
	}`)

		resp, err := client.Update(t.Context(), "bucket name", data)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("errors if PUT returns conflict", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: activeBucketResponse,
					}
				},
			},
			{
				PUT: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusConflict,
						ResponseBody: `some conflicting error'`,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "bucket name", []byte("{}"))

		assert.Zero(t, resp)

		var clientErr api.ClientError
		assert.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.MethodPut, clientErr.Operation)
		assert.Equal(t, "buckets", clientErr.Resource)
		assert.Equal(t, "bucket name", clientErr.Identifier)

		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusConflict, apiErr.StatusCode)
	})

	t.Run("errors if update payload is invalid JSON", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: activeBucketResponse,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.Update(t.Context(), "bucket name", []byte("{invalid"))

		assert.Zero(t, resp)

		var runtimeErr api.RuntimeError
		assert.ErrorAs(t, err, &runtimeErr)
		assert.Equal(t, "buckets", runtimeErr.Resource)
		assert.Equal(t, "bucket name", runtimeErr.Identifier)
		assert.Equal(t, "failed to unmarshal request payload", runtimeErr.Reason)
	})
}

func TestDecodingBucketResponses(t *testing.T) {
	const bucket1 = `{
	"bucketName": "bucket name",
	"table": "metrics",
	"displayName": "Default metrics (15 months)",
	"status": "active",
	"retentionDays": 462,
	"metricInterval": "PT1M",
	"version": 1
}`
	const bucket2 = `{
	"bucketName": "another name",
	"table": "metrics",
	"displayName": "Some logs",
	"status": "active",
	"retentionDays": 31,
	"metricInterval": "PT2M",
	"version": 42
}`

	type bucket struct {
		Name          string `json:"bucketName"`
		Table         string `json:"table"`
		RetentionDays int    `json:"retentionDays"`
	}

	t.Run("Get single bucket", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: bucket1,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())
		require.NoError(t, err)
		b, err := api.DecodeJSON[bucket](resp[0].Response)
		assert.NoError(t, err)

		assert.Equal(t, "bucket name", b.Name)
		assert.Equal(t, "metrics", b.Table)
		assert.Equal(t, 462, b.RetentionDays)
	})

	t.Run("List multiple buckets", func(t *testing.T) {
		payload := fmt.Sprintf(`{
		"buckets": [
			%s,
			%s
		]
	}`, bucket1, bucket2)

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, req *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: payload,
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		resp, err := client.List(t.Context())
		require.NoError(t, err)

		list, err := api.DecodeJSONObjects[bucket](resp[0])
		assert.NoError(t, err)
		assert.Len(t, list, 2)
		assert.Equal(t, bucket{
			Name:          "bucket name",
			Table:         "metrics",
			RetentionDays: 462,
		}, list[0])
		assert.Equal(t, bucket{
			Name:          "another name",
			Table:         "metrics",
			RetentionDays: 31,
		}, list[1])
	})
}
