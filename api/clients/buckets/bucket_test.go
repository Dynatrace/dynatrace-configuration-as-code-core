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
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/buckets"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
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

func TestGet(t *testing.T) {
	t.Run("successfully fetch a bucket", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
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

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Get(ctx, "bucket name")
		assert.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, body, []byte(activeBucketResponse))
	})

	t.Run("correctly create the error in case of a server issue", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
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

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Get(ctx, "bucket name")
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
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
				GET: func(t *testing.T, request *http.Request) testutils.Response {
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

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.List(ctx)
		assert.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, []byte(payload), body)
	})

	t.Run("successfully returns empty response if no buckets exist", func(t *testing.T) {
		const payload = `{ "buckets": [] }`
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
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

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.List(ctx)
		assert.NoError(t, err, "expected err to be nil")
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, []byte(payload), body)
	})

	t.Run("successfully returns response in case of HTTP error", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
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

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.List(ctx)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("returns error in case of network error", func(t *testing.T) {

		server := testutils.NewHTTPTestServer(t, nil)
		defer server.Close()

		faultyClient := server.FaultyClient()

		client := buckets.NewClient(rest.NewClient(server.URL(), faultyClient))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.List(ctx)
		assert.Error(t, err)
		assert.Empty(t, resp)
	})
}

func TestDelete(t *testing.T) {

	t.Run("delete bucket - OK", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusAccepted,
						ResponseBody: deletingBucketResponse,
					}
				},
			},
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: deletingBucketResponse,
					}
				},
			},
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
						ResponseBody: "",
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Delete(ctx, "bucket name")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, deletingBucketResponse, string(body))
	})

	t.Run("delete bucket - not found", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Delete(ctx, "bucket name")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, []byte{}, body)
	})

	t.Run("delete bucket - network error", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				DELETE: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusNotFound,
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Delete(ctx, "bucket name")
		assert.Error(t, err)
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

	t.Run("create bucket - OK", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				POST: func(t *testing.T, request *http.Request) testutils.Response {
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

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Create(ctx, []byte(someBucketData))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, creatingBucketResponse, string(body))
	})

	t.Run("create bucket - network error", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			// no request should reach test server
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Create(ctx, []byte(someBucketData))
		assert.Error(t, err)
		assert.Zero(t, resp)
	})
}

func TestUpdate(t *testing.T) {

	t.Run("update fails", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusForbidden,
						ResponseBody: "no write access message",
					}
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), &http.Client{}))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "bucket name", "1", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("update bucket - OK", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				PUT: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: updatingBucketResponse,
					}
				},
				ValidateRequest: func(t *testing.T, req *http.Request) {
					data, err := io.ReadAll(req.Body)
					assert.NoError(t, err)

					m := map[string]any{}
					err = json.Unmarshal(data, &m)
					assert.NoError(t, err)
				},
			},
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), &http.Client{}))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "bucket name", "1", data)
		assert.NoError(t, err)

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		m := map[string]any{}
		err = json.Unmarshal(body, &m)
		assert.NoError(t, err)
	})
}
