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

package openpipeline

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"net/http"
	"net/url"
)

const openPipelineResourcePath = "/platform/openpipeline/v1/configurations"

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
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}

	path, err := url.JoinPath(openPipelineResourcePath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := c.client.GET(ctx, path, ro)
	if err != nil {
		return nil, fmt.Errorf("unable to get object with ID %s: %w", id, err)
	}

	return resp, err
}

func (c Client) List(ctx context.Context, ro rest.RequestOptions) (*http.Response, error) {
	resp, err := c.client.GET(ctx, openPipelineResourcePath, ro)
	if err != nil {
		return nil, fmt.Errorf("unable to get objects: %w", err)
	}

	return resp, err
}

func (c Client) Update(ctx context.Context, id string, data []byte, ro rest.RequestOptions) (*http.Response, error) { //nolint:dupl
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}

	path, err := url.JoinPath(openPipelineResourcePath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := c.client.PUT(ctx, path, bytes.NewReader(data), ro)
	if err != nil {
		return nil, fmt.Errorf("unable to update object: %w", err)
	}
	return resp, err
}
