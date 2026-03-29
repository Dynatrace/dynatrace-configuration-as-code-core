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

package buckets

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	endpointPath = "/platform/storage/management/v1/bucket-definitions"
	resource     = "buckets"
)

var (
	idValidationErr       = api.ValidationError{Resource: resource, Field: "bucketName", Reason: "is empty"}
	defaultRequestOptions = rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests}
)

type bucketResponse struct {
	BucketName string `json:"bucketName"`
	Status     string `json:"status"`
	Version    int    `json:"version"`
}

// ListResponse is a Bucket API response containing multiple bucket objects.
type ListResponse = api.PagedListResponse

type listResponse struct {
	Buckets []json.RawMessage `json:"buckets"`
}

// Client is used to interact with the Grail bucket management API.
type Client struct {
	restClient *rest.Client
}

// NewClient creates a new buckets Client using the given rest.Client.
func NewClient(client *rest.Client) *Client {
	return &Client{restClient: client}
}

// Get retrieves a bucket definition by bucketName.
func (c Client) Get(ctx context.Context, bucketName string) (api.Response, error) {
	if bucketName == "" {
		return api.Response{}, idValidationErr
	}

	path, err := url.JoinPath(endpointPath, bucketName)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: bucketName, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.GET(ctx, path, defaultRequestOptions)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: bucketName, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: bucketName, Operation: http.MethodGet, Wrapped: err}
	}
	return resp, nil
}

// List retrieves all bucket definitions.
func (c Client) List(ctx context.Context) (ListResponse, error) {
	httpResp, err := c.restClient.GET(ctx, endpointPath, defaultRequestOptions)
	if err != nil {
		return ListResponse{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	apiResp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return ListResponse{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	r := listResponse{}
	if err = json.Unmarshal(apiResp.Data, &r); err != nil {
		return ListResponse{}, api.RuntimeError{Resource: resource, Reason: "unmarshalling failed", Wrapped: err}
	}

	b := make([][]byte, len(r.Buckets))
	for i, v := range r.Buckets {
		b[i], _ = v.MarshalJSON() // marshalling the JSON back to JSON will not fail
	}

	return ListResponse{
		api.ListResponse{
			Response: apiResp,
			Objects:  b,
		},
	}, nil
}

// Create sends a request to create a new bucket with the provided bucketName and data.
func (c Client) Create(ctx context.Context, bucketName string, data []byte) (api.Response, error) {
	if err := setBucketName(bucketName, &data); err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: bucketName, Reason: "failed to set bucket name in payload", Wrapped: err}
	}

	httpResp, err := c.restClient.POST(ctx, endpointPath, bytes.NewReader(data), defaultRequestOptions)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: bucketName, Operation: http.MethodPost, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: bucketName, Operation: http.MethodPost, Wrapped: err}
	}
	return resp, nil
}

// Delete sends a request to delete a bucket definition identified by bucketName.
func (c Client) Delete(ctx context.Context, bucketName string) (api.Response, error) {
	if bucketName == "" {
		return api.Response{}, idValidationErr
	}

	path, err := url.JoinPath(endpointPath, bucketName)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: bucketName, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.DELETE(ctx, path, defaultRequestOptions)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: bucketName, Operation: http.MethodDelete, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: bucketName, Operation: http.MethodDelete, Wrapped: err}
	}
	return resp, nil
}

// Update updates a bucket's data. If the data is unchanged compared to the existing bucket,
// no HTTP call is made and a 200 response is returned directly.
func (c Client) Update(ctx context.Context, bucketName string, data []byte) (api.Response, error) {
	if bucketName == "" {
		return api.Response{}, idValidationErr
	}

	apiResp, err := c.Get(ctx, bucketName)
	if err != nil {
		return api.Response{}, err
	}

	res, err := unmarshalJSON(apiResp.Data)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: bucketName, Reason: "failed to unmarshal GET response", Wrapped: err}
	}

	if bucketsEqual(apiResp.Data, data) {
		slog.DebugContext(ctx, "Configuration unmodified, no need to update bucket", slog.String("bucketName", bucketName))
		return api.Response{
			StatusCode: http.StatusOK,
		}, nil
	}

	var m map[string]interface{}
	if err = json.Unmarshal(data, &m); err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: bucketName, Reason: "failed to unmarshal request payload", Wrapped: err}
	}
	m["bucketName"] = res.BucketName
	m["version"] = res.Version
	m["status"] = res.Status

	data, err = json.Marshal(m)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: bucketName, Reason: "failed to marshal payload", Wrapped: err}
	}

	path, err := url.JoinPath(endpointPath, bucketName)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: bucketName, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.PUT(ctx, path, bytes.NewReader(data), rest.RequestOptions{
		QueryParams:           url.Values{"optimistic-locking-version": []string{strconv.Itoa(res.Version)}},
		CustomShouldRetryFunc: defaultRequestOptions.CustomShouldRetryFunc,
	})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: bucketName, Operation: http.MethodPut, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: bucketName, Operation: http.MethodPut, Wrapped: err}
	}
	return resp, nil
}

// bucketsEqual checks whether two bucket JSONs are equal in terms of update API calls.
// bucketName, version and status are assumed to be those of the existing object.
func bucketsEqual(existingBucket, newBucket []byte) bool {
	var existingMap map[string]interface{}
	if err := json.Unmarshal(existingBucket, &existingMap); err != nil {
		return false
	}

	var newMap map[string]interface{}
	if err := json.Unmarshal(newBucket, &newMap); err != nil {
		return false
	}

	// version and status are always taken from existing bucket on update
	newMap["bucketName"] = existingMap["bucketName"]
	newMap["version"] = existingMap["version"]
	newMap["status"] = existingMap["status"]

	return reflect.DeepEqual(existingMap, newMap)
}

// setBucketName sets the bucket name in the provided JSON data.
func setBucketName(bucketName string, data *[]byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(*data, &m); err != nil {
		return err
	}
	m["bucketName"] = bucketName
	var err error
	*data, err = json.Marshal(m)
	return err
}

// unmarshalJSON unmarshals JSON data into a bucketResponse struct.
func unmarshalJSON(body []byte) (bucketResponse, error) {
	r := bucketResponse{}
	if err := json.Unmarshal(body, &r); err != nil {
		return bucketResponse{}, err
	}
	return r, nil
}
