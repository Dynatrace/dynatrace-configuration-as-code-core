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
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/google/go-cmp/cmp"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/go-logr/logr"
)

const endpointPath = "platform/storage/management/v1/bucket-definitions"

type Response struct {
	api.Response
}

type response struct {
	api.Response
	BucketName string `json:"bucketName"`
	Status     string `json:"status"`
	Version    int    `json:"version"`
}

// ListResponse is a Bucket API response containing multiple bucket objects.
// For convenience, it contains a slice of Buckets in addition to the base api.Response data.
type ListResponse struct {
	api.ListResponse
}

type listResponse struct {
	api.Response
	Buckets []json.RawMessage `json:"buckets"`
}

type retrySettings struct {
	maxRetries           int
	durationBetweenTries time.Duration
	maxWaitDuration      time.Duration
}

type Client struct {
	client        *rest.Client
	retrySettings retrySettings
}

// Option represents a functional Option for the Client.
type Option func(*Client)

// WithRetrySettings sets the maximum number of retries as well as duration between retries.
// These settings are honored wherever retries are used in the Client - most notably in Client.Update and Client.Upsert,
// as well as Client.Create when waiting for a bucket to become available after creation.
//
// Parameters:
//   - maxRetries: maximum amount actions may be retries. (Some actions may ignore this and only honor maxWaitDuration)
//   - durationBetweenTries: time.Duration to wait between tries.
//   - maxWaitDuration: maximum time.Duration to wait before retrying is cancelled. If you supply a context.Context with a timeout, the shorter of the two will be honored.
func WithRetrySettings(maxRetries int, durationBetweenTries time.Duration, maxWaitDuration time.Duration) Option {
	return func(c *Client) {
		c.retrySettings = retrySettings{
			maxRetries:           maxRetries,
			durationBetweenTries: durationBetweenTries,
			maxWaitDuration:      maxWaitDuration,
		}
	}
}

// NewClient creates a new instance of a Client, which provides methods for interacting with the Grail bucket management API.
// This function initializes and returns a new Client instance that can be used to perform various operations
// on the remote server.
//
// Parameters:
//   - client: A pointer to a rest.Client instance used for making HTTP requests to the remote server.
//   - option: A variadic slice of client Option. Each Option will be applied to the new Client and define options such as retry settings.
//
// Returns:
//   - *Client: A pointer to a new Client instance initialized with the provided rest.Client and logger.
func NewClient(client *rest.Client, option ...Option) *Client {
	c := &Client{
		client: client,
		retrySettings: retrySettings{
			maxRetries:           5,
			durationBetweenTries: time.Second,
			maxWaitDuration:      2 * time.Minute,
		},
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
func (c Client) Get(ctx context.Context, bucketName string) (Response, error) {
	resp, err := c.get(ctx, bucketName)
	if err != nil {
		return Response{}, err
	}
	return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload, Request: resp.RequestInfo}}, nil
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
	resp, err := c.list(ctx)
	if err != nil {
		return ListResponse{}, err
	}

	b := make([][]byte, len(resp.Buckets))
	for i, v := range resp.Buckets {
		b[i], _ = v.MarshalJSON() // marshalling the JSON back to JSON will not fail
	}

	return ListResponse{
		ListResponse: api.ListResponse{
			Response: resp.Response,
			Objects:  b,
		},
	}, nil
}

// Create sends a request to the server to create a new bucket with the provided bucketName and data.
// The function prepares the data by setting the bucket name, then performs a POST request using the
// underlying client. It returns a Response and an error indicating the success or failure of its execution.
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
func (c Client) Create(ctx context.Context, bucketName string, data []byte) (Response, error) {
	resp, err := c.create(ctx, bucketName, data)
	if err != nil {
		return Response{}, err
	}
	return Response{
		Response: api.Response{
			StatusCode: resp.StatusCode,
			Data:       resp.Payload,
			Request:    resp.RequestInfo,
		},
	}, nil
}

// Update attempts to update a bucket's data using the provided client. It employs a retry mechanism
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
func (c Client) Update(ctx context.Context, bucketName string, data []byte) (Response, error) {

	logger := logr.FromContextOrDiscard(ctx)

	// check if bucket needs to be updated at all
	resp, err := c.get(ctx, bucketName)
	if err != nil {
		return Response{}, err
	}
	if bucketsEqual(resp.Payload, data) {
		logger.Info(fmt.Sprintf("Configuration unmodified, no need to update bucket with bucket name %q", bucketName))

		return Response{api.Response{
			StatusCode: 200,
		}}, nil
	}

	// attempt update
	ctx, cancel := context.WithTimeout(ctx, c.retrySettings.maxWaitDuration)
	defer cancel()
	for i := 0; i < c.retrySettings.maxRetries; i++ {
		select {
		case <-ctx.Done():
			return Response{}, fmt.Errorf("context cancelled before bucket with bucktName %q became available", bucketName)
		default:
			logger.V(1).Info(fmt.Sprintf("Trying to update bucket with bucket name %q (%d/%d retries)", bucketName, i+1, c.retrySettings.maxRetries))

			resp, err = c.getAndUpdate(ctx, bucketName, data)
			if err != nil {
				return Response{}, err
			}

			if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusBadRequest {
				return Response{api.Response{
					StatusCode: resp.StatusCode,
					Data:       resp.Payload,
					Request:    resp.RequestInfo,
				}}, nil
			}

			if resp.IsSuccess() {
				logger.Info(fmt.Sprintf("Updated bucket with bucket name %q", bucketName))
				return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload, Request: resp.RequestInfo}}, nil
			}
			time.Sleep(c.retrySettings.durationBetweenTries)
		}
	}
	return Response{
		Response: api.Response{
			StatusCode: resp.StatusCode,
			Data:       resp.Payload,
			Request:    resp.RequestInfo,
		},
	}, err
}

// Upsert creates or updates a bucket definition using the provided client. The function first attempts
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
func (c Client) Upsert(ctx context.Context, bucketName string, data []byte) (Response, error) {
	if bucketName == "" {
		return Response{}, fmt.Errorf("bucketName must be non-empty")
	}
	return c.upsert(ctx, bucketName, data)
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
func (c Client) Delete(ctx context.Context, bucketName string) (Response, error) {
	if bucketName == "" {
		return Response{}, fmt.Errorf("bucketName must be non-empty")
	}
	path, err := url.JoinPath(endpointPath, bucketName)
	if err != nil {
		return Response{}, fmt.Errorf("failed to create URL: %w", err)
	}
	resp, err := c.client.DELETE(ctx, path, rest.RequestOptions{})
	if err != nil {
		return Response{}, fmt.Errorf("unable to delete object with bucket name %q: %w", bucketName, err)
	}
	return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload, Request: resp.RequestInfo}}, err
}

// upsert is an internal function used by Upsert to perform the create or update logic.
func (c Client) upsert(ctx context.Context, bucketName string, data []byte) (Response, error) {

	logger := logr.FromContextOrDiscard(ctx)

	// First, try to create a new bucket definition
	resp, err := c.create(ctx, bucketName, data)
	if err != nil {
		return Response{}, err
	}

	// If creating the bucket definition worked, return the result
	if resp.IsSuccess() {
		logger.Info(fmt.Sprintf("Created bucket with bucket name %q", bucketName))
		return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload, Request: resp.RequestInfo}}, nil
	}

	// Return if creation failed, but the errors was not 409 Conflict - Bucket already exists
	if resp.StatusCode != http.StatusConflict {
		return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload, Request: resp.RequestInfo}}, err
	}

	// Otherwise, try to update an existing bucket definition
	logger.V(1).Info(fmt.Sprintf("Failed to create new object with bucket name %q. Trying to update existing object. API Error (HTTP %d): %s", bucketName, resp.StatusCode, resp.Payload))
	return c.Update(ctx, bucketName, data)
}

func (c Client) create(ctx context.Context, bucketName string, data []byte) (rest.Response, error) {
	if err := setBucketName(bucketName, &data); err != nil {
		return rest.Response{}, err
	}
	r, err := c.client.POST(ctx, endpointPath, bytes.NewReader(data), rest.RequestOptions{})
	if err != nil {
		return rest.Response{}, fmt.Errorf("failed to create object with bucketName %q: %w", bucketName, err)
	}
	if !r.IsSuccess() {
		return r, nil
	}

	timoutCtx, cancel := context.WithTimeout(ctx, c.retrySettings.maxWaitDuration)
	defer cancel() // cancel deadline if awaitBucketReady returns before deadline
	return c.awaitBucketReady(timoutCtx, bucketName)
}

func (c Client) awaitBucketReady(ctx context.Context, bucketName string) (rest.Response, error) {
	logger := logr.FromContextOrDiscard(ctx)

	for {
		select {
		case <-ctx.Done():
			return rest.Response{}, fmt.Errorf("context cancelled before bucket with bucktName %q became available", bucketName)
		default:
			// query bucket
			r, err := c.get(ctx, bucketName)
			if err != nil {
				return rest.Response{}, err
			}
			if !r.IsSuccess() && r.StatusCode != http.StatusNotFound { // if API returns 404 right after creation we want to wait
				return r, nil
			}

			// try to unmarshal into internal struct
			res, err := unmarshalJSON(&r)
			if err != nil {
				return r, err
			}

			if res.Status == "active" {
				logger.V(1).Info("Created bucket became active and ready to use")
				r.StatusCode = http.StatusCreated // return 'created' instead of the GET APIs 'ok'
				return r, nil
			}

			logger.V(1).Info("Waiting for bucket to become active after creation...")
			time.Sleep(c.retrySettings.durationBetweenTries)
		}
	}
}

func (c Client) get(ctx context.Context, bucketName string) (rest.Response, error) {
	path, err := url.JoinPath(endpointPath, bucketName)
	if err != nil {
		return rest.Response{}, fmt.Errorf("failed to create URL: %w", err)
	}
	return c.client.GET(ctx, path, rest.RequestOptions{})

}

func (c Client) list(ctx context.Context) (listResponse, error) {
	resp, err := c.client.GET(ctx, endpointPath, rest.RequestOptions{})
	if err != nil {
		return listResponse{}, fmt.Errorf("failed to list buckets:%w", err)
	}
	l, err := unmarshalJSONList(&resp)
	if err != nil {
		return listResponse{}, fmt.Errorf("failed to parse list response:%w", err)
	}
	return l, nil
}

func (c Client) getAndUpdate(ctx context.Context, bucketName string, data []byte) (rest.Response, error) {
	// try to get existing bucket definition
	b, err := c.get(ctx, bucketName)
	if err != nil {
		return rest.Response{}, fmt.Errorf("failed to get object with bucket name %q: %w", bucketName, err)
	}

	// return the result in case it's no HTTP 200
	if !b.IsSuccess() {
		return b, nil
	}

	// try to unmarshal into internal struct
	res, err := unmarshalJSON(&b)
	if err != nil {
		return rest.Response{}, err
	}

	// convert data to be sent to JSON
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return rest.Response{}, fmt.Errorf("unable to unmarshal template: %w", err)
	}
	m["bucketName"] = res.BucketName
	m["version"] = res.Version
	m["status"] = res.Status

	data, err = json.Marshal(m)
	if err != nil {
		return rest.Response{}, fmt.Errorf("unable to marshal data: %w", err)
	}

	// construct path for PUT request
	path, err := url.JoinPath(endpointPath, res.BucketName)
	if err != nil {
		return rest.Response{}, fmt.Errorf("failed to join URL: %w", err)
	}

	// make PUT request
	return c.client.PUT(ctx, path, bytes.NewReader(data), rest.RequestOptions{
		QueryParams: url.Values{"optimistic-locking-version": []string{strconv.Itoa(res.Version)}},
	})
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
		return err
	}
	m["bucketName"] = bucketName
	*data, err = json.Marshal(m)
	if err != nil {
		return err
	}
	return nil
}

// unmarshalJSON unmarshals JSON data into a response struct.
func unmarshalJSON(raw *rest.Response) (response, error) {
	var r response
	err := json.Unmarshal(raw.Payload, &r)
	if err != nil {
		return response{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	r.Data = raw.Payload
	r.StatusCode = raw.StatusCode
	return r, nil
}

// unmarshalJSONList unmarshals JSON data into a listResponse struct.
func unmarshalJSONList(raw *rest.Response) (listResponse, error) {
	var r listResponse
	err := json.Unmarshal(raw.Payload, &r)
	if err != nil {
		return listResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	r.Data = raw.Payload
	r.StatusCode = raw.StatusCode
	r.Request = raw.RequestInfo
	return r, nil
}
