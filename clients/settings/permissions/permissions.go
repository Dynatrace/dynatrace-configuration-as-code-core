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

package permissions

import (
	"bytes"
	"context"
	"net/url"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const endpointConfigPath = "platform/classic/environment-api/v2/settings/objects"
const permissionResourcePath = "permissions"
const allUsersAccessorType = "all-users"

type Client struct {
	client *rest.Client
}

func NewClient(client *rest.Client) *Client {
	return &Client{client: client}
}

func (c *Client) GetAllAccessors(ctx context.Context, objectID string) (api.Response, error) {
	return c.get(ctx, objectID, "", "")
}

func (c *Client) GetAllUsersAccessor(ctx context.Context, objectID string) (api.Response, error) {
	return c.get(ctx, objectID, allUsersAccessorType, "")
}

func (c *Client) GetAccessor(ctx context.Context, objectID string, accessorType string, accessorID string) (api.Response, error) {
	if accessorType == "" {
		return api.Response{}, ErrorPermissions{Wrapped: ErrorMissingAccessorType, Operation: GET}
	}

	if accessorID == "" {
		return api.Response{}, ErrorPermissions{Wrapped: ErrorMissingAccessorID, Operation: GET}
	}

	return c.get(ctx, objectID, accessorType, accessorID)
}

func (c *Client) get(ctx context.Context, objectID string, accessorType string, accessorID string) (api.Response, error) {
	if objectID == "" {
		return api.Response{}, ErrorPermissions{Wrapped: ErrorMissingObjectID, Operation: GET}
	}

	path, err := url.JoinPath(endpointConfigPath, objectID, permissionResourcePath, accessorType, accessorID)
	if err != nil {
		return api.Response{}, ErrorPermissions{Wrapped: err, Operation: GET}
	}

	resp, err := c.client.GET(ctx, path, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})

	if err != nil {
		return api.Response{}, ErrorPermissions{Wrapped: err, Operation: GET}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func (c *Client) Create(ctx context.Context, objectID string, body []byte) (api.Response, error) {
	if objectID == "" {
		return api.Response{}, ErrorPermissions{Wrapped: ErrorMissingObjectID, Operation: POST}
	}

	path, err := url.JoinPath(endpointConfigPath, objectID, permissionResourcePath)
	if err != nil {
		return api.Response{}, ErrorPermissions{Wrapped: err, Operation: POST}
	}

	resp, err := c.client.POST(ctx, path, bytes.NewReader(body), rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return api.Response{}, ErrorPermissions{Wrapped: err, Operation: POST}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func (c *Client) UpdateAllUsersAccessor(ctx context.Context, objectID string, body []byte) (api.Response, error) {
	return c.update(ctx, objectID, allUsersAccessorType, "", body)
}

func (c *Client) UpdateAccessor(ctx context.Context, objectID string, accessorType string, accessorID string, body []byte) (api.Response, error) {
	if accessorType == "" {
		return api.Response{}, ErrorPermissions{Wrapped: ErrorMissingAccessorType, Operation: PUT}
	}

	if accessorID == "" {
		return api.Response{}, ErrorPermissions{Wrapped: ErrorMissingAccessorID, Operation: PUT}
	}

	return c.update(ctx, objectID, accessorType, accessorID, body)
}

func (c *Client) update(ctx context.Context, objectID string, accessorType string, accessorID string, body []byte) (api.Response, error) {
	if objectID == "" {
		return api.Response{}, ErrorPermissions{Wrapped: ErrorMissingObjectID, Operation: PUT}
	}

	path, err := url.JoinPath(endpointConfigPath, objectID, permissionResourcePath, accessorType, accessorID)
	if err != nil {
		return api.Response{}, ErrorPermissions{Wrapped: err, Operation: PUT}
	}

	httpResp, err := c.client.PUT(ctx, path, bytes.NewReader(body), rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})

	if err != nil {
		return api.Response{}, ErrorPermissions{Wrapped: err, Operation: PUT}
	}

	return api.NewResponseFromHTTPResponse(httpResp)
}

func (c *Client) DeleteAllUsersAccessor(ctx context.Context, objectID string) (api.Response, error) {
	return c.delete(ctx, objectID, allUsersAccessorType, "")
}

func (c *Client) DeleteAccessor(ctx context.Context, objectID string, accessorType string, accessorID string) (api.Response, error) {
	if accessorType == "" {
		return api.Response{}, ErrorPermissions{Wrapped: ErrorMissingAccessorType, Operation: DELETE}
	}

	if accessorID == "" {
		return api.Response{}, ErrorPermissions{Wrapped: ErrorMissingAccessorID, Operation: DELETE}
	}

	return c.delete(ctx, objectID, accessorType, accessorID)
}

func (c *Client) delete(ctx context.Context, objectID string, accessorType string, accessorID string) (api.Response, error) {
	if objectID == "" {
		return api.Response{}, ErrorPermissions{Wrapped: ErrorMissingObjectID, Operation: DELETE}
	}

	path, err := url.JoinPath(endpointConfigPath, objectID, permissionResourcePath, accessorType, accessorID)
	if err != nil {
		return api.Response{}, ErrorPermissions{Wrapped: err, Operation: DELETE}
	}

	httpResp, err := c.client.DELETE(ctx, path, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})

	if err != nil {
		return api.Response{}, ErrorPermissions{Wrapped: err, Operation: DELETE}
	}

	return api.NewResponseFromHTTPResponse(httpResp)
}
