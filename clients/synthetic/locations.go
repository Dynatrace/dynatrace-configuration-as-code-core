// @license
// Copyright 2026 Dynatrace LLC
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

package synthetic

import (
	"bytes"
	"context"
	"net/http"
	"net/url"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	locationsPath             = basePath + "/locations"
	publicLocationsStatusPath = locationsPath + "/status"

	locationsResource             = "synthetic locations"
	publicLocationsStatusResource = "synthetic public locations status"
)

var locationsBasePath = url.URL{Path: locationsPath}

// ListLocationsOptions are the optional filters for ListLocations.
type ListLocationsOptions struct {
	// CloudPlatform filters by hosting platform (AWS, AZURE, ALIBABA, GOOGLE_CLOUD, OTHER).
	CloudPlatform string
	// Type filters by location type (PUBLIC, PRIVATE).
	Type string
	// Capability filters by supported capability (BROWSER, HTTP, HTTP_HIGH_RESOURCE, ICMP, TCP, DNS).
	Capability string
}

// ListLocations lists public and private synthetic locations.
func (c *Client) ListLocations(ctx context.Context, opts ListLocationsOptions) (api.ListResponse, error) {
	ro := rest.RequestOptions{}
	query := url.Values{}
	if opts.CloudPlatform != "" {
		query.Set("cloud-platform", opts.CloudPlatform)
	}
	if opts.Type != "" {
		query.Set("type", opts.Type)
	}
	if opts.Capability != "" {
		query.Set("capability", opts.Capability)
	}
	if len(query) > 0 {
		ro.QueryParams = query
	}

	resp, err := c.restClient.GET(ctx, locationsPath, ro)
	if err != nil {
		return api.ListResponse{}, api.ClientError{Resource: locationsResource, Operation: http.MethodGet, Wrapped: err}
	}

	return processListResponse(resp, locationsResource, "locations")
}

// GetLocation returns the location with the given id.
func (c *Client) GetLocation(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, api.ValidationError{Resource: locationsResource, Field: "id", Reason: "is empty"}
	}

	path := locationsBasePath.JoinPath(id).String()
	resp, err := c.restClient.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: locationsResource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// CreateLocation creates a private location.
func (c *Client) CreateLocation(ctx context.Context, body []byte) (api.Response, error) {
	resp, err := c.restClient.POST(ctx, locationsPath, bytes.NewReader(body), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: locationsResource, Operation: http.MethodPost, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// UpdateLocation updates the private location with the given id.
func (c *Client) UpdateLocation(ctx context.Context, id string, body []byte) (api.Response, error) {
	if id == "" {
		return api.Response{}, api.ValidationError{Resource: locationsResource, Field: "id", Reason: "is empty"}
	}

	path := locationsBasePath.JoinPath(id).String()
	resp, err := c.restClient.PUT(ctx, path, bytes.NewReader(body), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: locationsResource, Identifier: id, Operation: http.MethodPut, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// DeleteLocation deletes the private location with the given id.
func (c *Client) DeleteLocation(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, api.ValidationError{Resource: locationsResource, Field: "id", Reason: "is empty"}
	}

	path := locationsBasePath.JoinPath(id).String()
	resp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: locationsResource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// GetPublicLocationsStatus returns the status of public locations.
func (c *Client) GetPublicLocationsStatus(ctx context.Context) (api.Response, error) {
	resp, err := c.restClient.GET(ctx, publicLocationsStatusPath, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: publicLocationsStatusResource, Operation: http.MethodGet, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// UpdatePublicLocationsStatus updates the status of public locations.
func (c *Client) UpdatePublicLocationsStatus(ctx context.Context, body []byte) (api.Response, error) {
	resp, err := c.restClient.PUT(ctx, publicLocationsStatusPath, bytes.NewReader(body), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: publicLocationsStatusResource, Operation: http.MethodPut, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}
