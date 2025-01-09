// @license
// Copyright 2025 Dynatrace LLC
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

package slo

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const endpointPath = "platform/slo/v1/slos"

type Client struct {
	restClient *rest.Client
}

func NewClient(c *rest.Client) *Client {
	return &Client{
		restClient: c,
	}
}

func (c Client) List(ctx context.Context, ro rest.RequestOptions) (*http.Response, error) {
	path := endpointPath

	r, err := c.restClient.GET(ctx, path, ro)
	if err != nil {
		return nil, fmt.Errorf("failed to list segments: %w", err)
	}

	return r, nil
}

func (c Client) Get(ctx context.Context, id string, ro rest.RequestOptions) (*http.Response, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}

	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to join URL: %w", err)
	}

	r, err := c.restClient.GET(ctx, path, ro)
	if err != nil {
		return nil, fmt.Errorf("failed to get SLO with ID %s: %w", id, err)
	}

	return r, nil
}

func (c Client) Create(ctx context.Context, data []byte, ro rest.RequestOptions) (*http.Response, error) {
	r, err := c.restClient.POST(ctx, endpointPath, bytes.NewReader(data), ro)
	if err != nil {
		return nil, fmt.Errorf("failed to create new SLO: %w", err)
	}
	return r, nil
}

func (c Client) Update(ctx context.Context, id string, optimisticLockingVersion string, data []byte, ro rest.RequestOptions) (*http.Response, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}
	if optimisticLockingVersion == "" {
		return nil, fmt.Errorf("optimisticLockingVersion must be non-empty")
	}

	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to join URL: %w", err)
	}
	if ro.QueryParams == nil {
		ro.QueryParams = url.Values{}
	}
	ro.QueryParams.Add("optimistic-locking-version", optimisticLockingVersion)

	r, err := c.restClient.PUT(ctx, path, bytes.NewReader(data), ro)
	if err != nil {
		return nil, fmt.Errorf("failed to update SLO with ID %s: %w", id, err)
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

	r, err := c.restClient.DELETE(ctx, path, ro)
	if err != nil {
		return nil, fmt.Errorf("failed to delete SLO with ID %s: %w", id, err)
	}
	return r, nil
}
