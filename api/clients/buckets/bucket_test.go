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
	"context"
	"encoding/json"
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/buckets"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/internal/testutils"

	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
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

		responses := testutils.ServerResponses{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: payload,
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))

		resp, err := client.Get(context.TODO(), "bucket name")
		assert.NoError(t, err)
		assert.Equal(t, resp.Data, []byte(payload))
	})

	t.Run("correctly create the error in case of a server issue", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodGet: {
				ResponseCode: http.StatusNotFound,
				ResponseBody: "{}",
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))

		resp, err := client.Get(context.TODO(), "bucket name")
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

		responses := testutils.ServerResponses{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: payload,
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))

		resp, err := client.List(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, resp.Data, []byte(payload))
		assert.ElementsMatch(t, resp.Objects, [][]byte{[]byte(bucket1), []byte(bucket2)})
	})

	t.Run("successfully returns empty response if no buckets exist", func(t *testing.T) {
		const payload = `{ "buckets": [] }`
		responses := testutils.ServerResponses{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: payload,
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))

		resp, err := client.List(context.TODO())
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, resp.Data, []byte(payload))
		assert.Empty(t, resp.Objects)
	})

	t.Run("successfully returns response in case of HTTP error", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodGet: {
				ResponseCode: http.StatusNotFound,
				ResponseBody: "{}",
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))

		resp, err := client.List(context.TODO())
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("returns error in case of network error", func(t *testing.T) {

		server := testutils.NewHTTPTestServer(t, nil)
		defer server.Close()

		faultyClient := server.FaultyClient()

		client := buckets.NewClient(rest.NewClient(server.URL(), faultyClient, testr.New(t)), testr.New(t))

		resp, err := client.List(context.TODO())
		assert.Error(t, err)
		assert.Empty(t, resp)
	})
}

func TestUpsert(t *testing.T) {

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
		responses := testutils.ServerResponses{
			http.MethodPost: {
				ResponseCode: http.StatusBadRequest,
				ResponseBody: "ERROR",
			},
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: someBucketResponse,
			},
			http.MethodPut: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "no write access message",
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))

		data := []byte("{}")

		resp, err := client.Upsert(context.TODO(), "bucket name", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("create new bucket - OK", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodPost: {
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
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))
		data := []byte("{}")

		resp, err := client.Upsert(context.TODO(), "bucket name", data)
		assert.NoError(t, err)

		m := map[string]any{}
		err = json.Unmarshal(resp.Data, &m)
		assert.NoError(t, err)

		assert.Equal(t, "bucket name", m["bucketName"])
	})

	t.Run("update new bucket - OK", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodPost: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "this is an error",
			},
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
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))
		data := []byte("{}")

		resp, err := client.Upsert(context.TODO(), "bucket name", data)
		assert.NoError(t, err)

		m := map[string]any{}
		err = json.Unmarshal(resp.Data, &m)
		assert.NoError(t, err)

		assert.Equal(t, "bucket name", m["bucketName"])
	})

	t.Run("Update fails with conflict", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodPost: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "expected error, we don't want to create",
			},
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: someBucketResponse,
			},
			http.MethodPut: {
				ResponseCode: http.StatusConflict,
				ResponseBody: `some conflicting error'`,
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))
		data := []byte("{}")

		resp, err := client.Upsert(context.TODO(), "bucket name", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("Update fails because GET fails", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodPost: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "expected error, we don't want to create",
			},
			http.MethodGet: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "expected error, we don't want to get",
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))
		data := []byte("{}")

		resp, err := client.Upsert(context.TODO(), "bucket name", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	})

	t.Run("Update fails with conflict only once", func(t *testing.T) {
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
		client := buckets.NewClient(rest.NewClient(u, server.Client(), testr.New(t)), testr.New(t))
		data := []byte("{}")

		resp, err := client.Upsert(context.TODO(), "bucket name", data)
		assert.NoError(t, err)

		m := map[string]any{}
		err = json.Unmarshal(resp.Data, &m)
		assert.NoError(t, err)

		assert.Equal(t, "bucket name", m["bucketName"])
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
		responses := testutils.ServerResponses{
			http.MethodDelete: {
				ResponseCode: http.StatusAccepted,
				ResponseBody: someBucketResponse,
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))
		resp, err := client.Delete(context.TODO(), "bucket name")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
		assert.Equal(t, someBucketResponse, string(resp.Data))
	})

	t.Run("delete bucket - not found", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodDelete: {
				ResponseCode: http.StatusNotFound,
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))
		resp, err := client.Delete(context.TODO(), "bucket name")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Equal(t, []byte{}, resp.Data)
	})

	t.Run("delete bucket - network error", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodDelete: {
				ResponseCode: http.StatusNotFound,
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.FaultyClient(), testr.New(t)), testr.New(t))
		resp, err := client.Delete(context.TODO(), "bucket name")
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

	const someBucketResponse = `{
  "bucketName": "bucket name",
  "table": "metrics",
  "displayName": "Default metrics (15 months)",
  "status": "creating",
  "retentionDays": 462,
  "metricInterval": "PT1M",
  "version": 1
}`

	t.Run("create bucket - OK", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodPost: {
				ResponseCode: http.StatusCreated,
				ResponseBody: someBucketResponse,
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))
		resp, err := client.Create(context.TODO(), "bucket name", []byte(someBucketData))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, someBucketResponse, string(resp.Data))
	})

	t.Run("create bucket - network error", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodDelete: {
				ResponseCode: http.StatusNotFound,
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.FaultyClient(), testr.New(t)), testr.New(t))
		resp, err := client.Create(context.TODO(), "bucket name", []byte(someBucketData))
		assert.Error(t, err)
		assert.Zero(t, resp)
	})

	t.Run("create bucket - invalid data", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodDelete: {
				ResponseCode: http.StatusNotFound,
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))
		resp, err := client.Create(context.TODO(), "bucket name", []byte("-)§/$/(="))
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
		responses := testutils.ServerResponses{
			http.MethodPost: {
				ResponseCode: http.StatusBadRequest,
				ResponseBody: "ERROR",
			},
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: someBucketResponse,
			},
			http.MethodPut: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "no write access message",
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), &http.Client{}, testr.New(t)), testr.New(t))

		data := []byte("{}")

		resp, err := client.Upsert(context.TODO(), "bucket name", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("update new bucket - OK", func(t *testing.T) {
		responses := testutils.ServerResponses{
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
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), &http.Client{}, testr.New(t)), testr.New(t))
		data := []byte("{}")

		resp, err := client.Update(context.TODO(), "bucket name", data)
		assert.NoError(t, err)

		m := map[string]any{}
		err = json.Unmarshal(resp.Data, &m)
		assert.NoError(t, err)

		assert.Equal(t, "bucket name", m["bucketName"])
	})

	t.Run("Update fails with conflict", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: someBucketResponse,
			},
			http.MethodPut: {
				ResponseCode: http.StatusConflict,
				ResponseBody: `some conflicting error'`,
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), &http.Client{}, testr.New(t)), testr.New(t))
		data := []byte("{}")

		resp, err := client.Update(context.TODO(), "bucket name", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("Update fails because GET fails", func(t *testing.T) {
		responses := testutils.ServerResponses{
			http.MethodPost: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "expected error, we don't want to create",
			},
			http.MethodGet: {
				ResponseCode: http.StatusForbidden,
				ResponseBody: "expected error, we don't want to get",
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), &http.Client{}, testr.New(t)), testr.New(t))
		data := []byte("{}")

		resp, err := client.Update(context.TODO(), "bucket name", data)
		assert.NoError(t, err, "expected err to be nil")
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	})

	t.Run("Update fails with conflict only once", func(t *testing.T) {
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
		client := buckets.NewClient(rest.NewClient(u, &http.Client{}, testr.New(t)), testr.New(t))
		data := []byte("{}")

		resp, err := client.Update(context.TODO(), "bucket name", data)
		assert.NoError(t, err)

		m := map[string]any{}
		err = json.Unmarshal(resp.Data, &m)
		assert.NoError(t, err)

		assert.Equal(t, "bucket name", m["bucketName"])
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
		responses := testutils.ServerResponses{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: bucket1,
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))

		resp, err := client.List(context.TODO())
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

		responses := testutils.ServerResponses{
			http.MethodGet: {
				ResponseCode: http.StatusOK,
				ResponseBody: payload,
			},
		}
		server := testutils.NewHTTPTestServer(t, responses)
		defer server.Close()

		client := buckets.NewClient(rest.NewClient(server.URL(), server.Client(), testr.New(t)), testr.New(t))

		resp, err := client.List(context.TODO())
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
