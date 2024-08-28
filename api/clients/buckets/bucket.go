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

// Package buckets (api/clients/buckets) provides a simple CRUD client for the Grail Bucket API.
// For a 'smart' API client see package clients/buckets.
package buckets

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"net/http"
	"net/url"
)

const endpointPath = "platform/storage/management/v1/bucket-definitions"

type Client struct {
	client *rest.Client
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
func NewClient(client *rest.Client) *Client {
	client.SetHeader("Cache-Control", "no-cache")

	c := &Client{
		client: client,
	}

	return c
}

// Create a new bucket by making a HTTP POST request against the API.
//
// Parameters:
//   - ctx: Context for controlling the HTTP operation's lifecycle. Possibly containing a logger created with logr.NewContext.
//   - data: The data containing information about the new bucket.
//
// Returns:
//   - Response: A Response containing the result of the HTTP call, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (c Client) Create(ctx context.Context, data []byte) (*http.Response, error) {
	r, err := c.client.POST(ctx, endpointPath, bytes.NewReader(data), rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return nil, fmt.Errorf("failed to create new bucket: %w", err)
	}

	return r, nil
}

// Get a bucket by name, by making a HTTP GET /<bucketName> request against the API.
//
// Parameters:
//   - ctx: Context for controlling the HTTP operation's lifecycle. Possibly containing a logger created with logr.NewContext.
//   - bucketName: The name of the bucket to query.
//
// Returns:
//   - Response: A Response containing the result of the HTTP call, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (c Client) Get(ctx context.Context, bucketName string) (*http.Response, error) {
	path, err := url.JoinPath(endpointPath, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}
	return c.client.GET(ctx, path, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
}

// List all buckets, by making a HTTP GET request against the API.
//
// Parameters:
//   - ctx: Context for controlling the HTTP operation's lifecycle. Possibly containing a logger created with logr.NewContext.
//
// Returns:
//   - Response: A Response containing the result of the HTTP call, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (c Client) List(ctx context.Context) (*http.Response, error) {
	return c.client.GET(ctx, endpointPath, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
}

// Update a bucket by name, by making a HTTP PUT /<bucketName> request against the API.
//
// Parameters:
//   - ctx: Context for controlling the HTTP operation's lifecycle. Possibly containing a logger created with logr.NewContext.
//   - bucketName: The name of the bucket to update.
//   - bucketVersion: The expected version of the bucket. If this does not match the current version, the API will return a HTTP 409 Conflict error.
//   - data: The data to update the bucket to.
//
// Returns:
//   - Response: A Response containing the result of the HTTP call, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (c Client) Update(ctx context.Context, bucketName string, bucketVersion string, data []byte) (*http.Response, error) {
	path, err := url.JoinPath(endpointPath, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to join URL: %w", err)
	}

	return c.client.PUT(ctx, path, bytes.NewReader(data), rest.RequestOptions{
		QueryParams:           url.Values{"optimistic-locking-version": []string{bucketVersion}},
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
	})
}

// Delete a bucket by name, by making a HTTP DELETE /<bucketName> request against the API.
//
// Parameters:
//   - ctx: Context for controlling the HTTP operation's lifecycle. Possibly containing a logger created with logr.NewContext.
//   - bucketName: The name of the bucket to delete.
//
// Returns:
//   - Response: A Response containing the result of the HTTP call, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (c Client) Delete(ctx context.Context, bucketName string) (*http.Response, error) {
	if bucketName == "" {
		return nil, fmt.Errorf("bucketName must be non-empty")
	}
	path, err := url.JoinPath(endpointPath, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}
	return c.client.DELETE(ctx, path, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
}
