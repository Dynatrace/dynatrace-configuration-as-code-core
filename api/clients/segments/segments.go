// @license
// Copyright 2024 Dynatrace LLC
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

package segments

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const endpointPath = "platform/storage/filter-segments/v1/filter-segments"

type Client struct {
	client *rest.Client
}

func NewClient(client *rest.Client) *Client {
	c := &Client{
		client: client,
	}
	return c
}

// List gets the list of the existing segments. It uses the ":lean" endpoint to get the minimum necessary data set.
// If the field 'add-fields' in [rest.RequestOptions.QueryParams] is not specified, it will be set to "EXTERNALID".
func (c Client) List(ctx context.Context, ro rest.RequestOptions) (*http.Response, error) {
	path := endpointPath + ":lean" // minimal set of information is enough

	// set default behavior to request for ExternalID
	if !ro.QueryParams.Has("add-fields") {
		ro.QueryParams = url.Values{"add-fields": []string{"EXTERNALID"}}
	}

	r, err := c.client.GET(ctx, path, ro)
	if err != nil {
		return nil, fmt.Errorf("failed to list segments: %w", err)
	}

	return r, nil
}

// Get segment with UID.
// If the field 'add-fields' in [rest.RequestOptions.QueryParams] is not specified, it will be set to "INCLUDES", "VARIABLES", "EXTERNALID", "RESOURCECONTEXT".
func (c Client) Get(ctx context.Context, id string, ro rest.RequestOptions) (*http.Response, error) {
	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to join URL: %w", err)
	}

	// set default behavior to pick request for all information
	if !ro.QueryParams.Has("add-fields") {
		ro.QueryParams = url.Values{"add-fields": []string{"INCLUDES", "VARIABLES", "EXTERNALID", "RESOURCECONTEXT"}}
	}

	r, err := c.client.GET(ctx, path, ro)
	if err != nil {
		return nil, fmt.Errorf("failed to get segment with ID %s: %w", id, err)
	}

	return r, nil
}

func (c Client) Create(ctx context.Context, data []byte, ro rest.RequestOptions) (*http.Response, error) {
	r, err := c.client.POST(ctx, endpointPath, bytes.NewReader(data), ro)
	if err != nil {
		return nil, fmt.Errorf("failed to create new segment: %w", err)
	}
	return r, nil
}

func (c Client) Update(ctx context.Context, id string, data []byte, ro rest.RequestOptions) (*http.Response, error) {
	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to join URL: %w", err)
	}

	r, err := c.client.PUT(ctx, path, bytes.NewReader(data), ro)
	if err != nil {
		return nil, fmt.Errorf("failed to update segment with ID %s: %w", id, err)
	}
	return r, nil
}

func (c Client) Delete(ctx context.Context, id string, ro rest.RequestOptions) (*http.Response, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}

	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to join URL: %w", err)
	}

	r, err := c.client.DELETE(ctx, path, ro)
	if err != nil {
		return nil, fmt.Errorf("failed to delete segment with ID %s: %w", id, err)
	}
	return r, nil
}
