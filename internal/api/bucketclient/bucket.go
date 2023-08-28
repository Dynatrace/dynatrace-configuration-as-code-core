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

package bucketclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/internal/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/internal/rest"
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

type Client struct {
	client *rest.Client
	logger logr.Logger
}

// New creates a new instance of a Client, which provides methods for interacting with the Grail bucket management API.
// This function initializes and returns a new Client instance that can be used to perform various operations
// on the remote server.
//
// Parameters:
//   - client: A pointer to a rest.Client instance used for making HTTP requests to the remote server.
//   - logger: A logr.Logger instance used for logging relevant information during client operations.
//
// Returns:
//   - *Client: A pointer to a new Client instance initialized with the provided rest.Client and logger.
//
// New creates a new instance of the Client.
func New(client *rest.Client, logger logr.Logger) *Client {
	return &Client{
		client: client,
		logger: logger,
	}
}

// Get retrieves a bucket definition based on the provided bucketName. The function sends a GET request
// to the server using the given context and bucketName. It returns a Response and an error indicating
// the success or failure its execution.
//
// If the HTTP request to the server fails, the method returns an empty Response and an error explaining the issue.
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
	return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload}}, nil
}

// Create sends a request to the server to create a new bucket with the provided bucketName and data.
// The function prepares the data by setting the bucket name, then performs a POST request using the
// underlying client. It returns a Response and an error indicating the success or failure of its execution.
//
// If setting the bucket name in the data encounters an error, or if the HTTP request to the server
// fails, the function returns an empty Response and an error explaining the issue.
//
// Parameters:
//   - ctx: Context for controlling the HTTP operation's lifecycle.
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
// Parameters:
//   - ctx: Context for controlling the HTTP operation's lifecycle.
//   - bucketName: The name of the bucket to be updated.
//   - data: The new data to be assigned to the bucket.
//
// Returns:
//   - Response: A Response containing the result of the HTTP operation, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (c Client) Update(ctx context.Context, bucketName string, data []byte) (Response, error) {
	maxRetries := 3
	waitDuration := time.Second

	var resp rest.Response
	var err error
	for i := 0; i < maxRetries; i++ {
		c.logger.V(1).Info(fmt.Sprintf("Trying to update bucket with bucket name %q (%d/%d retries)", bucketName, i+1, maxRetries))

		resp, err = c.getAndUpdate(ctx, bucketName, data)
		if err != nil {
			return Response{}, err
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusBadRequest {
			return Response{api.Response{
				StatusCode: resp.StatusCode,
				Data:       resp.Payload,
			}}, nil
		}

		if resp.IsSuccess() {
			c.logger.Info(fmt.Sprintf("Updated bucket with bucket name %q", bucketName))
			return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload}}, nil
		}
		time.Sleep(waitDuration)
	}
	return Response{
		Response: api.Response{
			StatusCode: resp.StatusCode,
			Data:       resp.Payload,
		},
	}, err
}

// Upsert creates or updates a bucket definition using the provided client. The function first attempts
// to create the bucket. If the creation is successful, it returns the created bucket. If the creation
// fails, the function fetches the existing bucket and performs an update.
//
// In cases where the server does not immediately recognize an existing object after creation, retrying the GET
// request multiple times.
//
// If any HTTP request to the server fails, the method returns an empty Response and an error explaining the issue.
//
// Parameters:
//   - ctx: Context for controlling the upsert operation's lifecycle.
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
// Parameters:
//   - ctx: Context for controlling the deletion operation's lifecycle.
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
	return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload}}, err
}

// upsert is an internal function used by Upsert to perform the create or update logic.
func (c Client) upsert(ctx context.Context, bucketName string, data []byte) (Response, error) {
	// First, try to create a new bucket definition
	resp, err := c.create(ctx, bucketName, data)
	if err != nil {
		return Response{}, err
	}

	// If creating the bucket definition worked, return the result
	if resp.IsSuccess() {
		c.logger.Info(fmt.Sprintf("Created bucket with bucket name %q", bucketName))
		return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload}}, nil
	}

	// Otherwise, try to update an existing bucket definition
	c.logger.V(1).Info(fmt.Sprintf("Failed to create new object with bucket name %q. Trying to update existing object. Error: %s", bucketName, err))
	maxRetries := 3
	waitDuration := time.Second
	for i := 0; i < maxRetries; i++ {
		c.logger.V(1).Info(fmt.Sprintf("Trying to update bucket with bucket name %q (%d/%d retries)", bucketName, i+1, maxRetries))

		// Attempt to get and update the bucket's data
		resp, err = c.getAndUpdate(ctx, bucketName, data)
		if err != nil {
			return Response{}, err
		}

		// Check for specific HTTP status codes for early exits
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusBadRequest {
			return Response{api.Response{
				StatusCode: resp.StatusCode,
				Data:       resp.Payload,
			}}, nil
		}

		if resp.IsSuccess() {
			// Update operation was successful
			c.logger.Info(fmt.Sprintf("Updated bucket with bucket name %q", bucketName))
			return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload}}, nil
		}

		time.Sleep(waitDuration)
	}

	// All retries failed, return the last Response and error
	return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload}}, err
}

func (c Client) create(ctx context.Context, bucketName string, data []byte) (rest.Response, error) {
	if err := setBucketName(bucketName, &data); err != nil {
		return rest.Response{}, err
	}
	r, err := c.client.POST(ctx, endpointPath, bytes.NewReader(data), rest.RequestOptions{})
	if err != nil {
		return rest.Response{}, fmt.Errorf("failed to create object with bucketName %q: %w", bucketName, err)
	}
	return r, nil
}

func (c Client) get(ctx context.Context, bucketName string) (rest.Response, error) {
	path, err := url.JoinPath(endpointPath, bucketName)
	if err != nil {
		return rest.Response{}, fmt.Errorf("failed to create URL: %w", err)
	}
	return c.client.GET(ctx, path, rest.RequestOptions{})

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

	// construct path for PUT request
	path, err := url.JoinPath(endpointPath, res.BucketName)
	if err != nil {
		return rest.Response{}, fmt.Errorf("failed to join URL: %w", err)
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

	// make PUT request
	return c.client.PUT(ctx, path, bytes.NewReader(data), rest.RequestOptions{
		QueryParams: url.Values{"optimistic-locking-version": []string{strconv.Itoa(res.Version)}},
	})
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

// unmarshalJSON unmarshals JSON data into a Response struct.
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
