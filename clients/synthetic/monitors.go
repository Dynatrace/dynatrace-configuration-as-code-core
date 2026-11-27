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
	monitorsPath     = basePath + "/monitors"
	monitorsResource = "synthetic monitors"
)

var monitorsBasePath = url.URL{Path: monitorsPath}

// ListMonitorsOptions are the optional filters for ListMonitors.
type ListMonitorsOptions struct {
	// MonitorSelector narrows down the result, e.g. "externalId(my-id)" or "type(HTTP),enabled(true)".
	// It is sent in the query string, so very large selectors (e.g. many externalIds) can hit
	// URL length limits; callers should keep selectors small.
	MonitorSelector string
}

// ListMonitors lists synthetic monitors.
// The returned objects are summaries, use GetMonitor for the full payload.
func (c *Client) ListMonitors(ctx context.Context, opts ListMonitorsOptions) (api.ListResponse, error) {
	ro := rest.RequestOptions{}
	if opts.MonitorSelector != "" {
		ro.QueryParams = url.Values{"monitor-selector": {opts.MonitorSelector}}
	}

	resp, err := c.restClient.GET(ctx, monitorsPath, ro)
	if err != nil {
		return api.ListResponse{}, api.ClientError{Resource: monitorsResource, Operation: http.MethodGet, Wrapped: err}
	}

	return processListResponse(resp, monitorsResource, "monitors")
}

// GetMonitor returns the monitor with the given id.
func (c *Client) GetMonitor(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, api.ValidationError{Resource: monitorsResource, Field: "id", Reason: "is empty"}
	}

	path := monitorsBasePath.JoinPath(id).String()
	resp, err := c.restClient.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: monitorsResource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// CreateMonitor creates a monitor from the given JSON body.
func (c *Client) CreateMonitor(ctx context.Context, body []byte) (api.Response, error) {
	resp, err := c.restClient.POST(ctx, monitorsPath, bytes.NewReader(body), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: monitorsResource, Operation: http.MethodPost, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// UpdateMonitor updates the monitor with the given id.
func (c *Client) UpdateMonitor(ctx context.Context, id string, body []byte) (api.Response, error) {
	if id == "" {
		return api.Response{}, api.ValidationError{Resource: monitorsResource, Field: "id", Reason: "is empty"}
	}

	path := monitorsBasePath.JoinPath(id).String()
	resp, err := c.restClient.PUT(ctx, path, bytes.NewReader(body), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: monitorsResource, Identifier: id, Operation: http.MethodPut, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// DeleteMonitor deletes the monitor with the given id.
func (c *Client) DeleteMonitor(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, api.ValidationError{Resource: monitorsResource, Field: "id", Reason: "is empty"}
	}

	path := monitorsBasePath.JoinPath(id).String()
	resp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: monitorsResource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}
