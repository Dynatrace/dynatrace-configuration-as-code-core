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
	"net/http"
	"net/url"
	"strconv"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	endpointConfigPath     = "platform/classic/environment-api/v2/settings/objects"
	permissionResourcePath = "permissions"
	allUsersAccessorType   = "all-users"
	resource               = "permissions"
)

var (
	objectIDValidationErr     = api.ValidationError{Resource: resource, Field: "objectID", Reason: "is empty"}
	accessorTypeValidationErr = api.ValidationError{Resource: resource, Field: "accessorType", Reason: "is empty"}
	accessorIDValidationErr   = api.ValidationError{Resource: resource, Field: "accessorID", Reason: "is empty"}
)

// Client is used to interact with the Settings Permissions API.
type Client struct {
	restClient *rest.Client
}

// NewClient creates a new permissions Client using the given rest.Client.
func NewClient(client *rest.Client) *Client {
	return &Client{restClient: client}
}

// GetAllAccessors returns all accessors for a given settings object.
func (c Client) GetAllAccessors(ctx context.Context, objectID string, adminAccess bool) (api.Response, error) {
	return c.get(ctx, objectID, "", "", adminAccess)
}

// GetAllUsersAccessor returns the all-users accessor for a given settings object.
func (c Client) GetAllUsersAccessor(ctx context.Context, objectID string, adminAccess bool) (api.Response, error) {
	return c.get(ctx, objectID, allUsersAccessorType, "", adminAccess)
}

// GetAccessor returns a specific accessor for a given settings object.
func (c Client) GetAccessor(ctx context.Context, objectID string, accessorType string, accessorID string, adminAccess bool) (api.Response, error) {
	if accessorType == "" {
		return api.Response{}, accessorTypeValidationErr
	}

	if accessorID == "" {
		return api.Response{}, accessorIDValidationErr
	}

	return c.get(ctx, objectID, accessorType, accessorID, adminAccess)
}

func (c Client) get(ctx context.Context, objectID string, accessorType string, accessorID string, adminAccess bool) (api.Response, error) {
	if objectID == "" {
		return api.Response{}, objectIDValidationErr
	}

	path, err := url.JoinPath(endpointConfigPath, objectID, permissionResourcePath, accessorType, accessorID)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: objectID, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.GET(ctx, path, getRequestOptions(adminAccess))
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: objectID, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: objectID, Operation: http.MethodGet, Wrapped: err}
	}
	return resp, nil
}

// Create creates a new permission entry for a given settings object.
func (c Client) Create(ctx context.Context, objectID string, adminAccess bool, body []byte) (api.Response, error) {
	if objectID == "" {
		return api.Response{}, objectIDValidationErr
	}

	path, err := url.JoinPath(endpointConfigPath, objectID, permissionResourcePath)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: objectID, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.POST(ctx, path, bytes.NewReader(body), getRequestOptions(adminAccess))
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: objectID, Operation: http.MethodPost, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: objectID, Operation: http.MethodPost, Wrapped: err}
	}
	return resp, nil
}

// UpdateAllUsersAccessor updates the all-users accessor for a given settings object.
func (c Client) UpdateAllUsersAccessor(ctx context.Context, objectID string, adminAccess bool, body []byte) (api.Response, error) {
	return c.update(ctx, objectID, allUsersAccessorType, "", adminAccess, body)
}

// UpdateAccessor updates a specific accessor for a given settings object.
func (c Client) UpdateAccessor(ctx context.Context, objectID string, accessorType string, accessorID string, adminAccess bool, body []byte) (api.Response, error) {
	if accessorType == "" {
		return api.Response{}, accessorTypeValidationErr
	}

	if accessorID == "" {
		return api.Response{}, accessorIDValidationErr
	}

	return c.update(ctx, objectID, accessorType, accessorID, adminAccess, body)
}

func (c Client) update(ctx context.Context, objectID string, accessorType string, accessorID string, adminAccess bool, body []byte) (api.Response, error) {
	if objectID == "" {
		return api.Response{}, objectIDValidationErr
	}

	path, err := url.JoinPath(endpointConfigPath, objectID, permissionResourcePath, accessorType, accessorID)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: objectID, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.PUT(ctx, path, bytes.NewReader(body), getRequestOptions(adminAccess))
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: objectID, Operation: http.MethodPut, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: objectID, Operation: http.MethodPut, Wrapped: err}
	}
	return resp, nil
}

// DeleteAllUsersAccessor deletes the all-users accessor for a given settings object.
func (c Client) DeleteAllUsersAccessor(ctx context.Context, objectID string, adminAccess bool) (api.Response, error) {
	return c.delete(ctx, objectID, allUsersAccessorType, "", adminAccess)
}

// DeleteAccessor deletes a specific accessor for a given settings object.
func (c Client) DeleteAccessor(ctx context.Context, objectID string, accessorType string, accessorID string, adminAccess bool) (api.Response, error) {
	if accessorType == "" {
		return api.Response{}, accessorTypeValidationErr
	}

	if accessorID == "" {
		return api.Response{}, accessorIDValidationErr
	}

	return c.delete(ctx, objectID, accessorType, accessorID, adminAccess)
}

func (c Client) delete(ctx context.Context, objectID string, accessorType string, accessorID string, adminAccess bool) (api.Response, error) {
	if objectID == "" {
		return api.Response{}, objectIDValidationErr
	}

	path, err := url.JoinPath(endpointConfigPath, objectID, permissionResourcePath, accessorType, accessorID)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: objectID, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.DELETE(ctx, path, getRequestOptions(adminAccess))
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: objectID, Operation: http.MethodDelete, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: objectID, Operation: http.MethodDelete, Wrapped: err}
	}
	return resp, nil
}

func getRequestOptions(adminAccess bool) rest.RequestOptions {
	return rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           url.Values{"adminAccess": []string{strconv.FormatBool(adminAccess)}},
	}
}
