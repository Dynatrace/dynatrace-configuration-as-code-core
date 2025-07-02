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
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

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
		assert.Equal(t, resp.Data, []byte(activeBucketResponse))
	})

	t.Run("returns an error in case of a server issue", func(t *testing.T) {

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

		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusNotFound, apiError.StatusCode)
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
		assert.Equal(t, resp[0].Data, []byte(payload))
		assert.ElementsMatch(t, resp.All(), [][]byte{[]byte(bucket1), []byte(bucket2)})
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
		assert.Equal(t, resp[0].Data, []byte(payload))
		assert.Empty(t, resp.All())
	})

	t.Run("returns error in case of HTTP error", func(t *testing.T) {
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

		assert.Len(t, resp, 0)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusNotFound, apiError.StatusCode)

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
	t.Run("fails if bucket name is empty", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer server.Close()
		url, _ := url.Parse(server.URL)

		client := buckets.NewClient(rest.NewClient(url, server.Client()))
		_, err := client.Delete(t.Context(), "")
		assert.ErrorIs(t, err, buckets.ErrBucketEmpty)
	})

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
		assert.Equal(t, deletingBucketResponse, string(resp.Data))
	})

	t.Run("returns an error when bucket to be deleted is not found", func(t *testing.T) {

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

		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusNotFound, apiError.StatusCode)
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

		resp, err := client.Create(ctx, "bucket name", []byte(someBucketData))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, creatingBucketResponse, string(resp.Data))
	})

	t.Run("create bucket - network error", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			// no request should reach test server
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Create(ctx, "bucket name", []byte(someBucketData))
		assert.Error(t, err)
		assert.Zero(t, resp)
	})

	t.Run("create bucket - invalid data", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			// no request should reach test server
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Create(ctx, "bucket name", []byte("-)§/$/(="))
		assert.Error(t, err)
		assert.Zero(t, resp)
	})
}

func TestUpdate(t *testing.T) {

	t.Run("returns an error when update fails", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: activeBucketResponse,
					}
				},
			},
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
		resp, err := client.Update(ctx, "bucket name", data)

		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusForbidden, apiError.StatusCode)
	})

	t.Run("update bucket - OK", func(t *testing.T) {
		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: activeBucketResponse,
					}
				},
				ValidateRequest: func(t *testing.T, req *http.Request) {
					assert.Contains(t, req.URL.String(), url.PathEscape("bucket name"))
				},
			},
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

					assert.Equal(t, "bucket name", m["bucketName"])
				},
			},
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

		client := buckets.NewClient(rest.NewClient(server.URL(), &http.Client{}))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "bucket name", data)
		assert.NoError(t, err)

		m := map[string]any{}
		err = json.Unmarshal(resp.Data, &m)
		assert.NoError(t, err)

		assert.Equal(t, "bucket name", m["bucketName"])
	})

	t.Run("unmodified bucket - nothing happens", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: activeBucketResponse,
					}
				},
				ValidateRequest: func(t *testing.T, req *http.Request) {
					assert.Contains(t, req.URL.String(), url.PathEscape("bucket name"))
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

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Update(ctx, "bucket name", data)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("returns an error when update fails with conflict", func(t *testing.T) {

		responses := []testutils.ResponseDef{}

		for i := 0; i < 5; i++ {
			get := testutils.ResponseDef{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusOK,
						ResponseBody: activeBucketResponse,
					}
				},
			}
			responses = append(responses, get)

			put := testutils.ResponseDef{
				PUT: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusConflict,
						ResponseBody: `some conflicting error'`,
					}
				},
			}
			responses = append(responses, put)
		}

		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(
			rest.NewClient(server.URL(), server.Client()))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "bucket name", data)

		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusConflict, apiError.StatusCode)
	})

	t.Run("returns an error when update fails because GET fails", func(t *testing.T) {

		responses := []testutils.ResponseDef{
			{
				GET: func(t *testing.T, request *http.Request) testutils.Response {
					return testutils.Response{
						ResponseCode: http.StatusForbidden,
						ResponseBody: "expected error, we don't want to get",
					}
				},
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), &http.Client{}))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Update(ctx, "bucket name", data)

		assert.Zero(t, resp)
		var apiError api.APIError
		assert.ErrorAs(t, err, &apiError)
		assert.Equal(t, http.StatusForbidden, apiError.StatusCode)
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
				GET: func(t *testing.T, request *http.Request) testutils.Response {
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

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.List(ctx)
		assert.NoError(t, err)
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
