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
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/buckets"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/internal/testutils"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	t.Run("successfully fetch a bucket", func(t *testing.T) {
		const payload = `{
 "bucketName": "bucket name",
 "table": "metrics",
 "displayName": "Default metrics (15 months)",
 "status": "active",
 "retentionDays": 462,
 "metricInterval": "PT1M",
 "version": 1
}`

		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: payload,
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Get(ctx, "bucket name")
		assert.NoError(t, err)
		assert.Equal(t, resp.Data, []byte(payload))
	})

	t.Run("correctly create the error in case of a server issue", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusNotFound,
				ResponseBody: "{}",
			},
		}}
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

		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: payload,
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.List(ctx)
		assert.NoError(t, err)
		assert.Equal(t, resp.Data, []byte(payload))
		assert.ElementsMatch(t, resp.Objects, [][]byte{[]byte(bucket1), []byte(bucket2)})
	})

	t.Run("successfully returns empty response if no buckets exist", func(t *testing.T) {
		const payload = `{ "buckets": [] }`
		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: payload,
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.List(ctx)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, resp.Data, []byte(payload))
		assert.Empty(t, resp.Objects)
	})

	t.Run("successfully returns response in case of HTTP error", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusNotFound,
				ResponseBody: "{}",
			},
		}}
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

func TestUpsert(t *testing.T) {

	const creatingBucketResponse = `{
  "bucketName": "bucket name",
  "table": "metrics",
  "displayName": "Default metrics (15 months)",
  "status": "creating",
  "retentionDays": 462,
  "metricInterval": "PT1M",
  "version": 1
}`

	const activeBucketResponse = `{
  "bucketName": "bucket name",
  "table": "metrics",
  "displayName": "Default metrics (15 months)",
  "status": "active",
  "retentionDays": 462,
  "metricInterval": "PT1M",
  "version": 1
}`

	t.Run("create new bucket - OK", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodPost: {
				ResponseCode: http.StatusOK,
				ResponseBody: creatingBucketResponse,
				ValidateRequestFunc: func(req *http.Request) {
					data, err := io.ReadAll(req.Body)
					assert.NoError(t, err)

					m := map[string]any{}
					err = json.Unmarshal(data, &m)
					assert.NoError(t, err)

					assert.Equal(t, "bucket name", m["bucketName"])
				},
			},
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: activeBucketResponse,
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Upsert(ctx, "bucket name", data)
		assert.NoError(t, err)

		m := map[string]any{}
		err = json.Unmarshal(resp.Data, &m)
		assert.NoError(t, err)

		assert.Equal(t, "bucket name", m["bucketName"])
	})

	t.Run("create bucket - awaits bucket becoming ready", func(t *testing.T) {
		getRequests := 0
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			switch req.Method {
			case http.MethodPost:
				rw.WriteHeader(http.StatusCreated)
				rw.Write([]byte(creatingBucketResponse))
			case http.MethodGet:
				rw.WriteHeader(http.StatusOK)
				if getRequests < 5 {
					rw.Write([]byte(creatingBucketResponse))
					getRequests++
				} else {
					rw.Write([]byte(activeBucketResponse))
				}
			default:
				t.Fatalf("unexpected %s request", req.Method)

			}
		}))

		defer server.Close()

		url, _ := url.Parse(server.URL) //nolint:errcheck

		client := buckets.NewClient(rest.NewClient(url, server.Client()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Create(ctx, "bucket name", []byte("{}"))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, activeBucketResponse, string(resp.Data))
		assert.Equal(t, 5, getRequests)
	})

	t.Run("create fails", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodPost: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "Bucket exists",
			},
			http.MethodGet: {
				ResponseCode: http.StatusNotFound,
				ResponseBody: "{}",
			},
			http.MethodPut: {
				ResponseCode: http.StatusOK,
				ResponseBody: activeBucketResponse,
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Upsert(ctx, "bucket name", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("bucket exists, update - OK", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodPost: {
				ResponseCode: http.StatusConflict,
				ResponseBody: "Bucket exists",
			},
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: activeBucketResponse,
				ValidateRequestFunc: func(req *http.Request) {
					assert.Contains(t, req.URL.String(), url.PathEscape("bucket name"))
				},
			},
			http.MethodPut: {
				ResponseCode: http.StatusOK,
				ResponseBody: activeBucketResponse,
				ValidateRequestFunc: func(req *http.Request) {
					data, err := io.ReadAll(req.Body)
					assert.NoError(t, err)

					m := map[string]any{}
					err = json.Unmarshal(data, &m)
					assert.NoError(t, err)

					assert.Equal(t, "bucket name", m["bucketName"])
				},
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Upsert(ctx, "bucket name", data)
		assert.NoError(t, err)

		m := map[string]any{}
		err = json.Unmarshal(resp.Data, &m)
		assert.NoError(t, err)

		assert.Equal(t, "bucket name", m["bucketName"])
	})

	t.Run("bucket exists, update fails", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodPost: {
				ResponseCode: http.StatusConflict,
				ResponseBody: "Bucket exists",
			},
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: activeBucketResponse,
			},
			http.MethodPut: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "no write access message",
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Upsert(ctx, "bucket name", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("bucket exists, update fails with conflict", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodPost: {
				ResponseCode: http.StatusConflict,
				ResponseBody: "Bucket exists",
			},
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: activeBucketResponse,
			},
			http.MethodPut: {
				ResponseCode: http.StatusConflict,
				ResponseBody: `some conflicting error'`,
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(
			rest.NewClient(server.URL(), server.Client()),
			buckets.WithRetrySettings(5, 0, time.Minute))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Upsert(ctx, "bucket name", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("bucket exists, update fails because GET fails", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodPost: {
				ResponseCode: http.StatusConflict,
				ResponseBody: "expected error, we don't want to create",
			},
			http.MethodGet: {
				ResponseCode: http.StatusNotFound,
				ResponseBody: "expected error, we don't want to get",
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(
			rest.NewClient(server.URL(), server.Client()),
			buckets.WithRetrySettings(5, 0, time.Minute))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Upsert(ctx, "bucket name", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	})

	t.Run("bucket exists, update succeeds after initial conflict", func(t *testing.T) {
		var firstTry = true
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			switch req.Method {
			case http.MethodPost:
				rw.WriteHeader(http.StatusConflict)
				rw.Write([]byte("no, this is an error"))
			case http.MethodGet:
				rw.Write([]byte(activeBucketResponse))
			case http.MethodPut:
				if firstTry {
					rw.WriteHeader(http.StatusConflict)
					rw.Write([]byte("conflict"))
					firstTry = false
				} else {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte(activeBucketResponse))
				}
			default:
				assert.Failf(t, "unexpected method %q", req.Method)
			}
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := buckets.NewClient(rest.NewClient(u, server.Client()))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Upsert(ctx, "bucket name", data)
		assert.NoError(t, err)

		m := map[string]any{}
		err = json.Unmarshal(resp.Data, &m)
		assert.NoError(t, err)

		assert.Equal(t, "bucket name", m["bucketName"])
	})

	t.Run("bucket exists, but is not modified, no update happens", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodPost: {
				ResponseCode: http.StatusConflict,
				ResponseBody: "Bucket exists",
			},
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: activeBucketResponse,
				ValidateRequestFunc: func(req *http.Request) {
					assert.Contains(t, req.URL.String(), url.PathEscape("bucket name"))
				},
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))
		data := []byte(activeBucketResponse)

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Upsert(ctx, "bucket name", data)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestDelete(t *testing.T) {

	const someBucketResponse = `{
  "bucketName": "bucket name",
  "table": "metrics",
  "displayName": "Default metrics (15 months)",
  "status": "deleting",
  "retentionDays": 462,
  "metricInterval": "PT1M",
  "version": 1
}`

	t.Run("delete bucket - OK", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodDelete: {
				ResponseCode: http.StatusAccepted,
				ResponseBody: someBucketResponse,
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Delete(ctx, "bucket name")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
		assert.Equal(t, someBucketResponse, string(resp.Data))
	})

	t.Run("delete bucket - not found", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodDelete: {
				ResponseCode: http.StatusNotFound,
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Delete(ctx, "bucket name")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Equal(t, []byte{}, resp.Data)
	})

	t.Run("delete bucket - network error", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodDelete: {
				ResponseCode: http.StatusNotFound,
			},
		}}
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

	const creatingBucketResponse = `{
  "bucketName": "bucket name",
  "table": "metrics",
  "displayName": "Default metrics (15 months)",
  "status": "creating",
  "retentionDays": 462,
  "metricInterval": "PT1M",
  "version": 1
}`

	const activeBucketResponse = `{
  "bucketName": "bucket name",
  "table": "metrics",
  "displayName": "Default metrics (15 months)",
  "status": "active",
  "retentionDays": 462,
  "metricInterval": "PT1M",
  "version": 1
}`

	t.Run("create bucket - OK", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodPost: {
				ResponseCode: http.StatusCreated,
				ResponseBody: creatingBucketResponse,
			},
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: activeBucketResponse,
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Create(ctx, "bucket name", []byte(someBucketData))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, activeBucketResponse, string(resp.Data))
	})

	t.Run("create bucket - awaits bucket becoming ready", func(t *testing.T) {
		getRequests := 0
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			switch req.Method {
			case http.MethodPost:
				rw.WriteHeader(http.StatusCreated)
				rw.Write([]byte(creatingBucketResponse))
			case http.MethodGet:
				rw.WriteHeader(http.StatusOK)
				if getRequests < 5 {
					rw.Write([]byte(creatingBucketResponse))
					getRequests++
				} else {
					rw.Write([]byte(activeBucketResponse))
				}
			default:
				t.Fatalf("unexpected %s request", req.Method)

			}
		}))

		defer server.Close()

		url, _ := url.Parse(server.URL) //nolint:errcheck

		client := buckets.NewClient(rest.NewClient(url, server.Client()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Create(ctx, "bucket name", []byte(someBucketData))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, activeBucketResponse, string(resp.Data))
		assert.Equal(t, 5, getRequests)
	})

	t.Run("create bucket - network error", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			// no request should reach test server
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.FaultyClient()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Create(ctx, "bucket name", []byte(someBucketData))
		assert.Error(t, err)
		assert.Zero(t, resp)
	})

	t.Run("create bucket - invalid data", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			// no request should reach test server
		}}
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

	const someBucketResponse = `{
  "bucketName": "bucket name",
  "table": "metrics",
  "displayName": "Default metrics (15 months)",
  "status": "active",
  "retentionDays": 462,
  "metricInterval": "PT1M",
  "version": 1
}`
	t.Run("update fails", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: someBucketResponse,
			},
			http.MethodPut: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "no write access message",
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), &http.Client{}))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "bucket name", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("update bucket - OK", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: someBucketResponse,
				ValidateRequestFunc: func(req *http.Request) {
					assert.Contains(t, req.URL.String(), url.PathEscape("bucket name"))
				},
			},
			http.MethodPut: {
				ResponseCode: http.StatusOK,
				ResponseBody: someBucketResponse,
				ValidateRequestFunc: func(req *http.Request) {
					data, err := io.ReadAll(req.Body)
					assert.NoError(t, err)

					m := map[string]any{}
					err = json.Unmarshal(data, &m)
					assert.NoError(t, err)

					assert.Equal(t, "bucket name", m["bucketName"])
				},
			},
		}}
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
		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: someBucketResponse,
				ValidateRequestFunc: func(req *http.Request) {
					assert.Contains(t, req.URL.String(), url.PathEscape("bucket name"))
				},
			},
		}}
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

	t.Run("Update fails with conflict", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: someBucketResponse,
			},
			http.MethodPut: {
				ResponseCode: http.StatusConflict,
				ResponseBody: `some conflicting error'`,
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(
			rest.NewClient(server.URL(), server.Client()),
			buckets.WithRetrySettings(5, 0, time.Minute))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)
		resp, err := client.Update(ctx, "bucket name", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("Update fails because GET fails", func(t *testing.T) {
		responses := []testutils.ServerResponses{{
			http.MethodPost: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "expected error, we don't want to create",
			},
			http.MethodGet: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "expected error, we don't want to get",
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), &http.Client{}))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Update(ctx, "bucket name", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	})

	t.Run("Update fails at first, but succeeds after retry", func(t *testing.T) {
		var firstTry = true
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			switch req.Method {
			case http.MethodPost:
				rw.WriteHeader(http.StatusForbidden)
				rw.Write([]byte("no, this is an error"))
			case http.MethodGet:
				rw.Write([]byte(someBucketResponse))
			case http.MethodPut:
				if firstTry {
					rw.WriteHeader(http.StatusConflict)
					rw.Write([]byte("conflict"))
					firstTry = false
				} else {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte(someBucketResponse))
				}
			default:
				assert.Failf(t, "unexpected method %q", req.Method)
			}
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := buckets.NewClient(rest.NewClient(u, &http.Client{}))
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.Update(ctx, "bucket name", data)
		assert.NoError(t, err)

		m := map[string]any{}
		err = json.Unmarshal(resp.Data, &m)
		assert.NoError(t, err)

		assert.Equal(t, "bucket name", m["bucketName"])
	})

	t.Run("Update honors retrySettings maxWaitDuration", func(t *testing.T) {
		var firstTry = true
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			switch req.Method {
			case http.MethodPost:
				rw.WriteHeader(http.StatusForbidden)
				rw.Write([]byte("no, this is an error"))
			case http.MethodGet:
				rw.Write([]byte(someBucketResponse))
			case http.MethodPut:
				if firstTry {
					rw.WriteHeader(http.StatusConflict)
					rw.Write([]byte("conflict"))
					firstTry = false
				} else {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte(someBucketResponse))
				}
			default:
				assert.Failf(t, "unexpected method %q", req.Method)
			}
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		client := buckets.NewClient(rest.NewClient(u, &http.Client{}),
			buckets.WithRetrySettings(5, 0, 0)) // maxWaitDuration should time out immediately
		data := []byte("{}")

		ctx := testutils.ContextWithLogger(t)
		_, err := client.Update(ctx, "bucket name", data)
		assert.ErrorContains(t, err, "canceled")
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
		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: bucket1,
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.List(ctx)
		assert.NoError(t, err)
		b, err := api.DecodeJSON[bucket](resp.Response)
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

		responses := []testutils.ServerResponses{{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: payload,
			},
		}}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client()))

		ctx := testutils.ContextWithLogger(t)

		resp, err := client.List(ctx)
		assert.NoError(t, err)

		list, err := api.DecodeJSONObjects[bucket](resp.ListResponse)
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
