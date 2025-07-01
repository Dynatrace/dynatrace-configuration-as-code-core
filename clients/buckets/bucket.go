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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	endpointPath    = "platform/storage/management/v1/bucket-definitions"
	errUnmarshalMsg = "failed to unmarshal JSON response: %w"
	errMsg          = "failed to %s bucket: %w"
	errMsgWithName  = "failed to %s bucket with name %s: %w"
	getOperation    = "get"
	updateOperation = "update"
	upsertOperation = "upsert"
	deleteOperation = "delete"
	createOperation = "create"
	listOperation   = "list"
)

var ErrBucketEmpty = fmt.Errorf("bucketName must be non-empty")

type bucketResponse struct {
	BucketName string `json:"bucketName"`
	Status     string `json:"status"`
	Version    int    `json:"version"`
}

// ListResponse is a Bucket API response containing multiple bucket objects.
// For convenience, it contains a slice of Buckets in addition to the base api.Response data.
type ListResponse = api.PagedListResponse

type listResponse struct {
	Buckets []json.RawMessage `json:"buckets"`
}

type Client struct {
	restClient *rest.Client
}

// Option represents a functional Option for the Client.
type Option func(*Client)

// NewClient creates a new instance of a Client, which provides methods for interacting with the Grail bucket management API.
// This function initializes and returns a new Client instance that can be used to perform various operations
// on the remote server.
//
// Parameters:
//   - client: A pointer to a rest.Client instance used for making HTTP requests to the remote server.
//   - option: A variadic slice of apiClient Option. Each Option will be applied to the new Client and define options such as retry settings.
//
// Returns:
//   - *Client: A pointer to a new Client instance initialized with the provided rest.Client and logger.
func NewClient(client *rest.Client, option ...Option) *Client {
	client.SetHeader("Cache-Control", "no-cache")
	c := &Client{
		restClient: client,
	}

	for _, o := range option {
		o(c)
	}

	return c
}

// Get retrieves a bucket definition based on the provided bucketName. The function sends a GET request
// to the server using the given context and bucketName. It returns a Response and an error indicating
// the success or failure its execution.
//
// If the HTTP request to the server fails, the method returns an empty Response and an error explaining the issue.
//
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
//
// Parameters:
//   - ctx: Context for controlling the HTTP operation's lifecycle.
//   - bucketName: The name of the bucket to be retrieved.
//
// Returns:
//   - Response: A Response containing the result of the HTTP call, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (c Client) Get(ctx context.Context, bucketName string) (api.Response, error) {
	path, err := url.JoinPath(endpointPath, bucketName)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, getOperation, bucketName, err)
	}

	resp, err := c.restClient.GET(ctx, path, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, getOperation, bucketName, err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// List retrieves all bucket definitions. The function sends a GET request
// to the server using the given context. It returns a slice of bucket Responses and an error indicating
// the success or failure its execution.
//
// If the HTTP request to the server fails, the method returns an empty slice and an error explaining the issue.
//
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
//
// Parameters:
//   - ctx: Context for controlling the HTTP operation's lifecycle. Possibly containing a logger created with logr.NewContext.
//
// Returns:
//   - []Response: A slice of bucket Response containing the individual buckets resulting from the HTTP call, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (c Client) List(ctx context.Context) (ListResponse, error) {
	resp, err := c.restClient.GET(ctx, endpointPath, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return ListResponse{}, fmt.Errorf(errMsg, listOperation, err)
	}

	apiResp, err := api.NewResponseFromHTTPResponse(resp)
	if err != nil {
		return ListResponse{}, err
	}
	r := listResponse{}
	err = json.Unmarshal(apiResp.Data, &r)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "Failed to unmarshal JSON response")
		return ListResponse{}, fmt.Errorf(errMsg, listOperation, fmt.Errorf(errUnmarshalMsg, err))
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

// Create sends a request to the server to create a new bucket with the provided bucketName and data.
// The function prepares the data by setting the bucket name, then performs a POST request using the
// underlying apiClient. It returns a Response and an error indicating the success or failure of its execution.
//
// If setting the bucket name in the data encounters an error, or if the HTTP request to the server
// fails, the function returns an empty Response and an error explaining the issue.
//
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
//
// Parameters:
//   - ctx: Context for controlling the HTTP operation's lifecycle. Possibly containing a logger created with logr.NewContext.
//   - bucketName: The name of the bucket to be created.
//   - data: The data containing information about the new bucket.
//
// Returns:
//   - Response: A Response containing the result of the HTTP call, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (c Client) Create(ctx context.Context, bucketName string, data []byte) (api.Response, error) {
	if err := setBucketName(bucketName, &data); err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, createOperation, bucketName, fmt.Errorf("unable to set bucket name: %w", err))
	}
	r, err := c.restClient.POST(ctx, endpointPath, bytes.NewReader(data), rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, createOperation, bucketName, err)
	}

	return api.NewResponseFromHTTPResponse(r)
}

// Update attempts to update a bucket's data using the provided apiClient. It employs a retry mechanism
// in case of transient errors. The function returns a Response along with an error indicating the
// success or failure of its execution.
//
// The update process is retried up to a fixed maximum number of times. If the update fails
// with certain HTTP status codes (401 Unauthorized, 403 Forbidden, 400 Bad Request), the function
// returns an appropriate Response immediately. If the update is successful, the function returns
// a Response indicating success, or if all retries fail, it returns a Response and the last
// encountered error, if any.
//
// If the data to update the bucket fully matches what is already configured on the target environment,
// Update will not make an HTTP call, as this would needlessly increase the buckets version.
// This is transparent to callers and a normal StatusCode 200 Response is returned.
//
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
//
// Parameters:
//   - ctx: Context for controlling the HTTP operation's lifecycle. Possibly containing a logger created with logr.NewContext.
//   - bucketName: The name of the bucket to be updated.
//   - data: The new data to be assigned to the bucket.
//
// Returns:
//   - Response: A Response containing the result of the HTTP operation, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (c Client) Update(ctx context.Context, bucketName string, data []byte) (api.Response, error) {
	return c.getAndUpdate(ctx, bucketName, data)
}

// Upsert creates or updates a bucket definition using the provided apiClient. The function first attempts
// to create the bucket. If the creation is successful, it returns the created bucket. If the creation
// fails with a 409 conflict, the function fetches the existing bucket and performs an Update.
//
// If the creation fails with any other HTTP status (e.g. missing authorization or invalid payload) the
// HTTP Response is returned immediately, as attempting an Update would likely just fail as well.
//
// If any HTTP request to the server fails, the method returns an empty Response and an error explaining the issue.
//
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
//
// Parameters:
//   - ctx: Context for controlling the upsert operation's lifecycle. Possibly containing a logger created with logr.NewContext.
//   - bucketName: The name of the bucket to be upserted.
//   - data: The data for creating or updating the bucket.
//
// Returns:
//   - Response: A Response containing the result of the HTTP call, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (c Client) Upsert(ctx context.Context, bucketName string, data []byte) (api.Response, error) {
	if bucketName == "" {
		return api.Response{}, fmt.Errorf(errMsg, upsertOperation, ErrBucketEmpty)
	}
	logger := logr.FromContextOrDiscard(ctx)

	// First, try to create a new bucket definition
	resp, err := c.Create(ctx, bucketName, data)
	// If creating the bucket definition worked, return the result
	if err == nil {
		logger.Info(fmt.Sprintf("Created bucket '%s'", bucketName))
		return resp, nil
	}

	apiErr := api.APIError{}
	if !errors.As(err, &apiErr) {
		return api.Response{}, fmt.Errorf(errMsgWithName, upsertOperation, bucketName, err)
	}

	// Return if creation failed, but the errors was not 409 Conflict - Bucket already exists
	if apiErr.StatusCode != http.StatusConflict {
		return api.Response{}, apiErr
	}

	// Try to update an existing bucket definition
	logger.V(1).Info(fmt.Sprintf("Failed to create bucket '%s'. Trying to update existing bucket definition. API Error (HTTP %d): %s", bucketName, apiErr.StatusCode, apiErr.Body))
	return c.Update(ctx, bucketName, data)
}

// Delete sends a request to the server to delete a bucket definition identified by the provided bucketName.
// It returns a Response and an error indicating the success or failure of the deletion operation.
//
// If the provided bucketName is empty, the function returns an error indicating that the bucketName must be non-empty.
// If the HTTP request to the server fails, the method returns an empty Response and an error explaining the issue.
//
// If you wish to receive logs from this method supply a logger inside the context using logr.NewContext.
//
// Parameters:
//   - ctx: Context for controlling the deletion operation's lifecycle. Possibly containing a logger created with logr.NewContext.
//   - bucketName: The name of the bucket to be deleted.
//
// Returns:
//   - Response: A Response containing the result of the HTTP call, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (c Client) Delete(ctx context.Context, bucketName string) (api.Response, error) {
	if bucketName == "" {
		return api.Response{}, fmt.Errorf(errMsg, deleteOperation, ErrBucketEmpty)
	}
	path, err := url.JoinPath(endpointPath, bucketName)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, deleteOperation, bucketName, err)
	}

	resp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})

	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, deleteOperation, bucketName, err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func (c Client) getAndUpdate(ctx context.Context, bucketName string, data []byte) (api.Response, error) {
	logger := logr.FromContextOrDiscard(ctx)
	// try to get existing bucket definition
	apiResp, err := c.Get(ctx, bucketName)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, updateOperation, bucketName, err)
	}

	// try to unmarshal into internal struct
	res, err := unmarshalJSON(apiResp.Data)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, updateOperation, bucketName, err)
	}

	if bucketsEqual(apiResp.Data, data) {
		logger.Info(fmt.Sprintf("Configuration unmodified, no need to update bucket '%s''", bucketName))

		return api.Response{
			StatusCode: 200,
		}, nil
	}

	// convert data to be sent to JSON
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, updateOperation, bucketName, fmt.Errorf("unable to unmarshal data: %w", err))
	}
	m["bucketName"] = res.BucketName
	m["version"] = res.Version
	m["status"] = res.Status

	data, err = json.Marshal(m)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, updateOperation, bucketName, fmt.Errorf("unable to marshal data: %w", err))
	}

	path, err := url.JoinPath(endpointPath, bucketName)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, updateOperation, bucketName, err)
	}

	resp, err := c.restClient.PUT(ctx, path, bytes.NewReader(data), rest.RequestOptions{
		QueryParams:           url.Values{"optimistic-locking-version": []string{strconv.Itoa(res.Version)}},
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
	})

	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, updateOperation, bucketName, err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// bucketsEqual checks whether two bucket JSONs are equal in terms of update API calls
// this means that like for an update bucketName, version and status are assumed to be
// those of the existing object, ignoring what ever may be defined in the supplied data.
func bucketsEqual(exists, new []byte) bool {
	var existsMap map[string]interface{}
	if err := json.Unmarshal(exists, &existsMap); err != nil {
		return false
	}

	var newMap map[string]interface{}
	if err := json.Unmarshal(new, &newMap); err != nil {
		return false
	}
	// version and status are always taken from existing bucket on update
	newMap["bucketName"] = existsMap["bucketName"]
	newMap["version"] = existsMap["version"]
	newMap["status"] = existsMap["status"]

	return cmp.Equal(existsMap, newMap)
}

// setBucketName sets the bucket name in the provided JSON data.
func setBucketName(bucketName string, data *[]byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(*data, &m)
	if err != nil {
		return fmt.Errorf("unable to unmarshal data: %w", err)
	}
	m["bucketName"] = bucketName
	*data, err = json.Marshal(m)
	if err != nil {
		return fmt.Errorf("unable to marshal data: %w", err)
	}
	return nil
}

// unmarshalJSON unmarshals JSON data into a response struct.
func unmarshalJSON(body []byte) (bucketResponse, error) {
	r := bucketResponse{}
	err := json.Unmarshal(body, &r)
	if err != nil {
		return bucketResponse{}, fmt.Errorf(errUnmarshalMsg, err)
	}
	return r, nil
}
