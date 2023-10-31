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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	bucketAPI "github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/buckets"
	"github.com/google/go-cmp/cmp"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/go-logr/logr"
)

const (
	stateActive   = "active"
	stateDeleting = "deleting"
)

type response struct {
	api.Response
	BucketName string `json:"bucketName"`
	Status     string `json:"status"`
	Version    int    `json:"version"`
}

type Response = api.Response

// ListResponse is a Bucket API response containing multiple bucket objects.
// For convenience, it contains a slice of Buckets in addition to the base api.Response data.
type ListResponse = api.PagedListResponse

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
	apiClient     *bucketAPI.Client
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
//   - maxWaitDuration: maximum time.Duration to wait before retrying is canceled. If you supply a context.Context with a timeout, the shorter of the two will be honored.
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
//   - option: A variadic slice of apiClient Option. Each Option will be applied to the new Client and define options such as retry settings.
//
// Returns:
//   - *Client: A pointer to a new Client instance initialized with the provided rest.Client and logger.
func NewClient(client *rest.Client, option ...Option) *Client {
	c := &Client{
		apiClient: bucketAPI.NewClient(client),
		retrySettings: retrySettings{
			maxRetries:           15,
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
	resp, err := c.apiClient.Get(ctx, bucketName)
	if err != nil {
		return api.Response{}, err
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return Response{}, err
	}
	return api.Response{StatusCode: resp.StatusCode, Data: body, Request: rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()}}, nil
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
	resp, err := c.apiClient.List(ctx)
	if err != nil {
		return ListResponse{}, fmt.Errorf("failed to list buckets:%w", err)
	}
	l, err := unmarshalJSONList(resp)
	if err != nil {
		return ListResponse{}, fmt.Errorf("failed to parse list response:%w", err)
	}

	b := make([][]byte, len(l.Buckets))
	for i, v := range l.Buckets {
		b[i], _ = v.MarshalJSON() // marshalling the JSON back to JSON will not fail
	}

	return ListResponse{
		api.ListResponse{
			Response: l.Response,
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
func (c Client) Create(ctx context.Context, bucketName string, data []byte) (Response, error) {
	resp, err := c.create(ctx, bucketName, data)
	if err != nil {
		return api.Response{}, err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return Response{}, err
	}

	return api.Response{
		StatusCode: resp.StatusCode,
		Data:       body,
		Request:    rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()},
	}, nil
}

var DeletingBucketErr = errors.New("cannot update bucket that is currently being deleted")

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
func (c Client) Update(ctx context.Context, bucketName string, data []byte) (Response, error) {

	logger := logr.FromContextOrDiscard(ctx)

	ctx, cancel := context.WithTimeout(ctx, c.retrySettings.maxWaitDuration)
	defer cancel()

	// get current state of the bucket
	resp, err := c.apiClient.Get(ctx, bucketName)
	if err != nil {
		return api.Response{}, err
	}
	if !rest.IsSuccess(resp) {
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return Response{}, err
		}
		r := api.Response{
			StatusCode: resp.StatusCode,
			Data:       body,
			Request: rest.RequestInfo{
				Method: "GET",
			},
		}
		return r, nil
	}

	current, err := unmarshalJSON(resp)
	if err != nil {
		return api.Response{}, err
	}

	if current.Status == stateDeleting {
		return api.Response{}, DeletingBucketErr
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return Response{}, err
	}

	if bucketsEqual(body, data) {
		logger.Info(fmt.Sprintf("Configuration unmodified, no need to update bucket with bucket name %q", bucketName))

		return api.Response{
			StatusCode: 200,
		}, nil
	}

	// attempt update
	if current.Status != stateActive {
		logger.V(1).Info(fmt.Sprintf("Waiting for bucket with bucket name %q to reach updatable state - currently %s", bucketName, current.Status))
		if _, err := c.awaitBucketState(ctx, bucketName, active); err != nil {
			return api.Response{}, err
		}
	}

	resp, err = c.getAndUpdate(ctx, bucketName, data)
	if err != nil {
		return api.Response{}, err
	}

	if rest.IsSuccess(resp) {
		logger.Info(fmt.Sprintf("Updated bucket with bucket name %q", bucketName))
	}

	return api.Response{StatusCode: resp.StatusCode, Data: body, Request: rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()}}, nil
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
func (c Client) Upsert(ctx context.Context, bucketName string, data []byte) (Response, error) {
	if bucketName == "" {
		return api.Response{}, fmt.Errorf("bucketName must be non-empty")
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
	resp, err := c.apiClient.Delete(ctx, bucketName)
	if err != nil {
		return api.Response{}, fmt.Errorf("unable to delete object with bucket name %q: %w", bucketName, err)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return Response{}, err
	}

	if !rest.IsSuccess(resp) {
		return api.Response{StatusCode: resp.StatusCode, Data: body, Request: rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()}}, nil
	}

	// await bucket being successfully deleted
	timeoutCtx, cancel := context.WithTimeout(ctx, c.retrySettings.maxWaitDuration)
	defer cancel() // cancel deadline if awaitBucketState returns before deadline
	_, err = c.awaitBucketState(timeoutCtx, bucketName, removed)
	if err != nil {
		return api.Response{}, fmt.Errorf("unable to delete object with bucket name %q: %w", bucketName, err)
	}

	return api.Response{StatusCode: resp.StatusCode, Data: body, Request: rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()}}, nil
}

// upsert is an internal function used by Upsert to perform the create or update logic.
func (c Client) upsert(ctx context.Context, bucketName string, data []byte) (Response, error) {

	logger := logr.FromContextOrDiscard(ctx)
	ctx, cancel := context.WithTimeout(ctx, c.retrySettings.maxWaitDuration)
	defer cancel()

	// First, try to create a new bucket definition
	resp, err := c.create(ctx, bucketName, data)
	if err != nil {
		return api.Response{}, err
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return Response{}, err
	}

	// If creating the bucket definition worked, return the result
	if rest.IsSuccess(resp) {
		logger.Info(fmt.Sprintf("Created bucket with bucket name %q", bucketName))
		return api.Response{StatusCode: resp.StatusCode, Data: body, Request: rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()}}, nil
	}

	// Return if creation failed, but the errors was not 409 Conflict - Bucket already exists
	if resp.StatusCode != http.StatusConflict {
		return api.Response{StatusCode: resp.StatusCode, Data: body, Request: rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()}}, err
	}

	// If bucket is currently being deleted, wait for it to be gone, then re-create it
	if b, err := unmarshalJSON(resp); err != nil && b.Status == "deleting" {
		logger.V(1).Info(fmt.Sprintf("Bucket %q is being deleted. Waiting before re-creation...", b.BucketName))
		if _, err := c.awaitBucketState(ctx, bucketName, removed); err != nil {
			return Response{}, err
		}
		return c.Create(ctx, bucketName, data)
	}

	// Try to update an existing bucket definition
	logger.V(1).Info(fmt.Sprintf("Failed to create new object with bucket name %q. Trying to update existing object. API Error (HTTP %d): %s", bucketName, resp.StatusCode, body))
	apiResp, err := c.Update(ctx, bucketName, data)

	if errors.Is(err, DeletingBucketErr) {
		logger.V(1).Info(fmt.Sprintf("Failed to upsert bucket with name %q as it was being deleted. Re-creating...", bucketName))
		if _, err := c.awaitBucketState(ctx, bucketName, removed); err != nil {
			return Response{}, err
		}
		return c.Create(ctx, bucketName, data)
	}
	return apiResp, err
}

func (c Client) create(ctx context.Context, bucketName string, data []byte) (*http.Response, error) {
	if err := setBucketName(bucketName, &data); err != nil {
		return nil, err
	}
	r, err := c.apiClient.Create(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create object with bucketName %q: %w", bucketName, err)
	}
	if !rest.IsSuccess(r) {
		return r, nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, c.retrySettings.maxWaitDuration)
	defer cancel() // cancel deadline if awaitBucketState returns before deadline
	return c.awaitBucketState(timeoutCtx, bucketName, active)
}

type bucketAvailability int

const (
	active bucketAvailability = iota
	removed
)

func (c Client) awaitBucketState(ctx context.Context, bucketName string, desired bucketAvailability) (*http.Response, error) {
	logger := logr.FromContextOrDiscard(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context canceled before bucket with bucktName %q became available", bucketName)
		default:
			// query bucket
			r, err := c.apiClient.Get(ctx, bucketName)
			if err != nil {
				return nil, err
			}
			if !rest.IsSuccess(r) && r.StatusCode != http.StatusNotFound { // if API returns 404 right after creation we want to wait
				return r, nil
			}

			switch desired {
			case active:
				// try to unmarshal into internal struct
				res, err := unmarshalJSON(r)
				if err != nil {
					return r, err
				}

				if res.Status == "active" {
					logger.V(1).Info("Bucket became active and ready to use")
					r.StatusCode = http.StatusCreated // return 'created' instead of the GET APIs 'ok'
					return r, nil
				}
			case removed:
				if r.StatusCode == http.StatusNotFound {
					logger.V(1).Info("Bucket was removed")
					return r, nil
				}
			}

			logger.V(1).Info("Waiting for bucket to reach desired state...")
			time.Sleep(c.retrySettings.durationBetweenTries)
		}
	}
}

func (c Client) getAndUpdate(ctx context.Context, bucketName string, data []byte) (*http.Response, error) {
	// try to get existing bucket definition
	b, err := c.apiClient.Get(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to get object with bucket name %q: %w", bucketName, err)
	}

	// return the result in case it's no HTTP 200
	if !rest.IsSuccess(b) {
		return b, nil
	}

	// try to unmarshal into internal struct
	res, err := unmarshalJSON(b)
	if err != nil {
		return nil, err
	}

	// convert data to be sent to JSON
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal template: %w", err)
	}
	m["bucketName"] = res.BucketName
	m["version"] = res.Version
	m["status"] = res.Status

	data, err = json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal data: %w", err)
	}

	resp, err := c.apiClient.Update(ctx, bucketName, strconv.Itoa(res.Version), data)

	if err != nil {
		return nil, err
	}
	if !rest.IsSuccess(resp) {
		return resp, nil
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, c.retrySettings.maxWaitDuration)
	defer cancel() // cancel deadline if awaitBucketState returns before deadline
	return c.awaitBucketState(timeoutCtx, bucketName, active)
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
func unmarshalJSON(raw *http.Response) (response, error) {
	var r response
	body, err := io.ReadAll(raw.Body)
	if err != nil {
		return r, err
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return response{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	r.Data = body
	r.StatusCode = raw.StatusCode
	return r, nil
}

// unmarshalJSONList unmarshals JSON data into a listResponse struct.
func unmarshalJSONList(raw *http.Response) (listResponse, error) {
	var r listResponse
	body, err := io.ReadAll(raw.Body)
	if err != nil {
		return r, err
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return listResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	r.Data = body
	r.StatusCode = raw.StatusCode
	r.Request = rest.RequestInfo{Method: raw.Request.Method, URL: raw.Request.URL.String()}
	return r, nil
}
