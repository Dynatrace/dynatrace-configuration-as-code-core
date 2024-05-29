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

package clients

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"net/http"
	"net/url"
)

type Client struct {
	client       *rest.Client
	resourcePath string
}

type RequestOptions struct {
	QueryParams  url.Values
	ContentType  string
	ResourcePath string
}

func NewClient(client *rest.Client, resourcePath string) *Client {
	c := &Client{
		client:       client,
		resourcePath: resourcePath,
	}
	return c
}

func (c Client) Get(ctx context.Context, id string, ro RequestOptions) (*http.Response, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}

	path, err := url.JoinPath(c.getResourcePath(ro), id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := c.client.GET(ctx, path, c.getRequestOptions(ro))
	if err != nil {
		return nil, fmt.Errorf("unable to get object with ID %s: %w", id, err)
	}

	return resp, err
}

func (c Client) Create(ctx context.Context, data []byte, ro RequestOptions) (*http.Response, error) {
	resp, err := c.client.POST(ctx, c.getResourcePath(ro), bytes.NewReader(data), c.getRequestOptions(ro))
	if err != nil {
		return nil, fmt.Errorf("unable to create object: %w", err)
	}
	return resp, err
}

func (c Client) List(ctx context.Context, ro RequestOptions) (*http.Response, error) {
	resp, err := c.client.GET(ctx, c.getResourcePath(ro), c.getRequestOptions(ro))
	if err != nil {
		return nil, fmt.Errorf("unable to get objects: %w", err)
	}

	return resp, err
}

func (c Client) Update(ctx context.Context, id string, data []byte, ro RequestOptions) (*http.Response, error) { //nolint:dupl
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}

	path, err := url.JoinPath(c.getResourcePath(ro), id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := c.client.PUT(ctx, path, bytes.NewReader(data), c.getRequestOptions(ro))
	if err != nil {
		return nil, fmt.Errorf("unable to update object: %w", err)
	}
	return resp, err
}

func (c Client) Patch(ctx context.Context, id string, data []byte, ro RequestOptions) (*http.Response, error) { //nolint:dupl
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}

	path, err := url.JoinPath(c.getResourcePath(ro), id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := c.client.PATCH(ctx, path, bytes.NewReader(data), c.getRequestOptions(ro))
	if err != nil {
		return nil, fmt.Errorf("unable to update object: %w", err)
	}
	return resp, err
}

func (c Client) Delete(ctx context.Context, id string, ro RequestOptions) (*http.Response, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}

	path, err := url.JoinPath(c.getResourcePath(ro), id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := c.client.DELETE(ctx, path, c.getRequestOptions(ro))
	if err != nil {
		return nil, fmt.Errorf("unable to delete object: %w", err)
	}
	return resp, err
}

func (c Client) getResourcePath(ro RequestOptions) string {
	resourcePath := c.resourcePath
	if ro.ResourcePath != "" {
		resourcePath = ro.ResourcePath
	}
	return resourcePath
}

func (c Client) getRequestOptions(ro RequestOptions) rest.RequestOptions {
	return rest.RequestOptions{
		QueryParams: ro.QueryParams,
		ContentType: ro.ContentType,
	}
}
