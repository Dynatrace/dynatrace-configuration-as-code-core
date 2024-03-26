/*
 * @license
 * Copyright 2023 Dynatrace LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package automation (api/clients/automation) provides a simple CRUD client for the Automations API.
// For a 'smart' API client see package clients/automation.
package automation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"net/http"
	"net/url"
	"strconv"
)

// Resource specifies information about a specific resource
type Resource struct {
	// Path is the API path to be used for this resource
	Path string
}

// ResourceType enumerates different kind of resources
type ResourceType int

const (
	Workflows ResourceType = iota
	BusinessCalendars
	SchedulingRules
)

var Resources = map[ResourceType]Resource{
	Workflows:         {Path: "/platform/automation/v1/workflows"},
	BusinessCalendars: {Path: "/platform/automation/v1/business-calendars"},
	SchedulingRules:   {Path: "/platform/automation/v1/scheduling-rules"},
}

// NewClient creates and returns a new instance a client which is used for interacting
// with automation resources.
//
// Parameters:
//
//   - client: A REST client used for making HTTP requests to interact with automation resources.
//
// Returns:
//
//   - *Client: A new instance of the Client type initialized with the provided REST client and resources.
func NewClient(client *rest.Client) *Client {
	c := &Client{
		client: client,
	}

	return c
}

// Client can be used to interact with the Automation API
type Client struct {
	client *rest.Client
}

// ClientOption are (optional) additional parameter passed to the creation of
// an automation client
type ClientOption func(*Client)

// Get retrieves a single automation object based on the specified resource type and ID.
// It makes a HTTP GET <resourceType endpoint>/<id> request against the API.
//
// It checks if the ID is non-empty, and if not, returns an error.
//
// Parameters:
//
//   - ctx: The context for the HTTP request.
//   - resourceType: The type of the resource to retrieve.
//   - id: The unique identifier of the object to retrieve.
//
// Returns:
//
//   - Response: A Response containing the result of the HTTP operation, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (a Client) Get(ctx context.Context, resourceType ResourceType, id string) (*http.Response, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be non empty")
	}
	path, err := url.JoinPath(Resources[resourceType].Path, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	return a.makeRequestWithAdminAccess(resourceType, func(options rest.RequestOptions) (*http.Response, error) {
		return a.client.GET(ctx, path, rest.RequestOptions{})
	})

}

// Create creates a new automation object based on the specified resource type.
//
// Parameters:
//
//   - ctx: The context for the HTTP request.
//   - resourceType: The type of the resource to retrieve.
//   - data: the data of the resource
//
// Returns:
//
//   - Response: A Response containing the result of the HTTP operation, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (a Client) Create(ctx context.Context, resourceType ResourceType, data []byte) (result *http.Response, err error) {
	return a.makeRequestWithAdminAccess(resourceType, func(options rest.RequestOptions) (*http.Response, error) {
		return a.client.POST(ctx, Resources[resourceType].Path, bytes.NewReader(data), options)
	})
}

// List retrieves makes a HTTP GET <resourceType endpoint>/ request against the API.
// The Automations API implements offset based pagination, which can be controlled by passing the desired offset to this
// method. Parsing responses and making several calls with the desired offset needs to be handled by the caller.
// See the 'smart' client implemented in clients/automation/Client for a List implementation that follows pagination automatically.
//
// Parameters:
//   - ctx: A context.Context for controlling the request lifecycle.
//   - resourceType: The type of resource to list.
//   - offset: The offset to start the paginated query from. This is passed directly to the API.
//
// Returns:
//
//   - ListResponse: A ListResponse which is an api.PagedListResponse containing all objects fetched from the api
//   - error: An error if the HTTP call fails or another error happened.
func (a Client) List(ctx context.Context, resourceType ResourceType, offset int) (*http.Response, error) {
	return a.makeRequestWithAdminAccess(resourceType, func(options rest.RequestOptions) (*http.Response, error) {
		options.QueryParams["offset"] = []string{strconv.Itoa(offset)}
		return a.client.GET(ctx, Resources[resourceType].Path, options)
	})
}

// Update updates an automation object based on the specified resource type and id
//
// Parameters:
//
//   - ctx: The context for the HTTP request.
//   - resourceType: The type of the resource to retrieve.
//   - id: the id of the resource to be updated
//   - data: the updated data
//
// Returns:
//
//   - Response: A Response containing the result of the HTTP operation, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (a Client) Update(ctx context.Context, resourceType ResourceType, id string, data []byte) (*http.Response, error) {
	if id == "" {
		return nil, errors.New("id must be non empty")
	}
	path, err := url.JoinPath(Resources[resourceType].Path, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	return a.makeRequestWithAdminAccess(resourceType, func(options rest.RequestOptions) (*http.Response, error) {
		return a.client.PUT(ctx, path, bytes.NewReader(data), options)
	})
}

// Delete removes an automation object of the specified resource type by its unique identifier (ID).
//
// If the initial DELETE request results in a forbidden status code (HTTP 403) for Workflows, it retries
// the request without the "adminAccess" parameter.
//
// Parameters:
//   - ctx: A context.Context for controlling the request lifecycle.
//   - resourceType: The type of resource from which to delete the object.
//   - id: The unique identifier (ID) of the object to delete.
//
// Returns:
//
//   - Response: A Response containing the result of the HTTP operation, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (a Client) Delete(ctx context.Context, resourceType ResourceType, id string) (*http.Response, error) {
	if id == "" {
		return nil, errors.New("id must be non empty")
	}
	path, err := url.JoinPath(Resources[resourceType].Path, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	return a.makeRequestWithAdminAccess(resourceType, func(options rest.RequestOptions) (*http.Response, error) {
		return a.client.DELETE(ctx, path, options)
	})
}

func (a Client) makeRequestWithAdminAccess(resourceType ResourceType, request func(options rest.RequestOptions) (*http.Response, error)) (*http.Response, error) {
	opt := rest.RequestOptions{
		QueryParams: url.Values{"adminAccess": []string{strconv.FormatBool(resourceType == Workflows)}},
	}

	resp, err := request(opt)
	if err != nil {
		return nil, err
	}

	// if Workflow API rejected the initial request with admin permissions -> retry without
	if resp != nil && resp.StatusCode == http.StatusForbidden {
		return request(rest.RequestOptions{})
	}

	return resp, err
}
