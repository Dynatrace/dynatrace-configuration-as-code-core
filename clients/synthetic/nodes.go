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
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	nodesPath     = basePath + "/nodes"
	nodesResource = "synthetic nodes"
)

var nodesBasePath = url.URL{Path: nodesPath}

// ListNodesOptions are the optional filters for ListNodes. A nil value means the filter is not applied.
type ListNodesOptions struct {
	// AssignedToLocation returns only nodes assigned to a location (true) or only unassigned nodes (false).
	AssignedToLocation *bool
	// Containerized returns only containerized nodes (true) or only non-containerized nodes (false).
	Containerized *bool
}

// ListNodes lists synthetic nodes.
func (c *Client) ListNodes(ctx context.Context, opts ListNodesOptions) (api.ListResponse, error) {
	ro := rest.RequestOptions{}
	query := url.Values{}
	if opts.AssignedToLocation != nil {
		query.Set("assigned-to-location", strconv.FormatBool(*opts.AssignedToLocation))
	}
	if opts.Containerized != nil {
		query.Set("containerized", strconv.FormatBool(*opts.Containerized))
	}
	if len(query) > 0 {
		ro.QueryParams = query
	}

	resp, err := c.restClient.GET(ctx, nodesPath, ro)
	if err != nil {
		return api.ListResponse{}, api.ClientError{Resource: nodesResource, Operation: http.MethodGet, Wrapped: err}
	}

	return processListResponse(resp, nodesResource, "nodes")
}

// GetNode returns the node with the given id.
func (c *Client) GetNode(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, api.ValidationError{Resource: nodesResource, Field: "id", Reason: "is empty"}
	}

	path := nodesBasePath.JoinPath(id).String()
	resp, err := c.restClient.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: nodesResource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}
