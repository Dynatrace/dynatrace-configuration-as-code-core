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

package grailfiltersegements

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const endpointPath = "platform/storage/filter-segments/v0.1/filter-segments"

type Client struct {
	client *rest.Client
}

func NewClient(client *rest.Client) *Client {
	c := &Client{
		client: client,
	}
	return c
}

func (c Client) Get(ctx context.Context, id string, ro rest.RequestOptions) (*http.Response, error) {
	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}
	// set to pick all information by default
	u, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}
	qp := u.Query()
	qp.Set("add-field", "INCLUDES")
	u.RawQuery = qp.Encode()
	return c.client.GET(ctx, u.String(), ro)
}

func (c Client) List(ctx context.Context) (*http.Response, error) {
	return c.client.GET(ctx, endpointPath, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
}

func (c Client) Create(ctx context.Context, data []byte) (*http.Response, error) {
	r, err := c.client.POST(ctx, endpointPath, bytes.NewReader(data), rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return nil, fmt.Errorf("failed to create new filter segment: %w", err)
	}
	return r, nil
}

func (c Client) Update(ctx context.Context, id string, data []byte, ro rest.RequestOptions) (*http.Response, error) {
	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to join URL: %w", err)
	}
	return c.client.PUT(ctx, path, bytes.NewReader(data), ro)
}

func (c Client) Delete(ctx context.Context, id string) (*http.Response, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}
	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}
	return c.client.DELETE(ctx, path, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
}
