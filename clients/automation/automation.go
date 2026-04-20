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

package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

// ResourceType enumerates the different kinds of automation resources.
type ResourceType int

const (
	Workflows ResourceType = iota
	BusinessCalendars
	SchedulingRules
)

var resources = map[ResourceType]string{
	Workflows:         "/platform/automation/v1/workflows",
	BusinessCalendars: "/platform/automation/v1/business-calendars",
	SchedulingRules:   "/platform/automation/v1/scheduling-rules",
}

func resourceName(rt ResourceType) string {
	switch rt {
	case Workflows:
		return "automation-workflow"
	case BusinessCalendars:
		return "automation-business-calendar"
	case SchedulingRules:
		return "automation-scheduling-rule"
	default:
		return "automation"
	}
}

func idValidationErr(rt ResourceType) api.ValidationError {
	return api.ValidationError{Resource: resourceName(rt), Field: "id", Reason: "is empty"}
}

// Client is used to interact with the Automation API.
type Client struct {
	restClient *rest.Client
}

// NewClient creates a new automation Client using the given rest.Client.
func NewClient(client *rest.Client) *Client {
	return &Client{restClient: client}
}

// Get returns one specific automation object by resource type and ID.
func (c Client) Get(ctx context.Context, resourceType ResourceType, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr(resourceType)
	}

	resource := resourceName(resourceType)

	path, err := url.JoinPath(resources[resourceType], id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.getWithAdminAccess(ctx, resourceType, path)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}
	return resp, nil
}

// Create creates a new automation object of the specified resource type.
func (c Client) Create(ctx context.Context, resourceType ResourceType, data []byte) (api.Response, error) {
	resource := resourceName(resourceType)

	httpResp, err := c.postWithAdminAccess(ctx, resourceType, resources[resourceType], data)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPost, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPost, Wrapped: err}
	}
	return resp, nil
}

// Update updates an existing automation object by resource type and ID.
func (c Client) Update(ctx context.Context, resourceType ResourceType, id string, data []byte) (api.Response, error) {
	resource := resourceName(resourceType)

	if err := removeIDField(&data); err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to remove id field from payload", Wrapped: err}
	}

	if id == "" {
		return api.Response{}, idValidationErr(resourceType)
	}

	path, err := url.JoinPath(resources[resourceType], id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.putWithAdminAccess(ctx, resourceType, path, data)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodPut, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodPut, Wrapped: err}
	}
	return resp, nil
}

// List returns all automation objects of the specified resource type.
func (c Client) List(ctx context.Context, resourceType ResourceType) (api.PagedListResponse, error) {
	var pagedListResponse api.PagedListResponse

	resource := resourceName(resourceType)
	totalCount := 1 // ensure starting condition is met
	retrieved := 0
	useAdminAccess := resourceType == Workflows

	for retrieved < totalCount {
		result, err := c.listPage(ctx, resourceType, resource, useAdminAccess, retrieved)
		if err != nil {
			return api.PagedListResponse{}, err
		}

		if useAdminAccess && !result.adminAccessUsed {
			useAdminAccess = false
			continue
		}

		totalCount = result.totalCount
		retrieved += len(result.Objects)
		pagedListResponse = append(pagedListResponse, result.ListResponse)
	}

	return pagedListResponse, nil
}

type listPageResult struct {
	api.ListResponse
	totalCount      int
	adminAccessUsed bool
}

func (c Client) listPage(ctx context.Context, resourceType ResourceType, resource string, useAdminAccess bool, offset int) (listPageResult, error) {
	opts := rest.RequestOptions{
		QueryParams: url.Values{
			"offset": []string{strconv.Itoa(offset)},
		},
	}
	if useAdminAccess {
		opts.QueryParams["adminAccess"] = []string{"true"}
	}

	httpResp, err := c.restClient.GET(ctx, resources[resourceType], opts)
	if err != nil {
		return listPageResult{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		var apiErr api.APIError
		// if Workflow API rejected the request with admin permissions -> retry without
		if useAdminAccess && isAPIError(err, &apiErr) && apiErr.StatusCode == http.StatusForbidden {
			return listPageResult{adminAccessUsed: false}, nil
		}
		return listPageResult{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	var listResp struct {
		Count   int               `json:"count"`
		Results []json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(resp.Data, &listResp); err != nil {
		return listPageResult{}, api.RuntimeError{Resource: resource, Reason: "unmarshalling failed", Wrapped: err}
	}

	objects := make([][]byte, len(listResp.Results))
	for i, v := range listResp.Results {
		objects[i] = v
	}

	return listPageResult{
		ListResponse: api.ListResponse{
			Response: api.Response{
				StatusCode: httpResp.StatusCode,
				Data:       resp.Data,
				Request:    api.NewRequestInfoFromRequest(httpResp.Request),
			},
			Objects: objects,
		},
		totalCount:      listResp.Count,
		adminAccessUsed: useAdminAccess,
	}, nil
}

// Delete removes an automation object of the specified resource type by ID.
func (c Client) Delete(ctx context.Context, resourceType ResourceType, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr(resourceType)
	}

	resource := resourceName(resourceType)

	path, err := url.JoinPath(resources[resourceType], id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.deleteWithAdminAccess(ctx, resourceType, path)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}
	return resp, nil
}

// getWithAdminAccess performs a GET request, trying with adminAccess=true for Workflows first.
// If the request is forbidden, it retries without adminAccess.
func (c Client) getWithAdminAccess(ctx context.Context, rt ResourceType, path string) (*http.Response, error) {
	return c.requestWithAdminAccess(rt, func(opts rest.RequestOptions) (*http.Response, error) {
		return c.restClient.GET(ctx, path, opts)
	})
}

// postWithAdminAccess performs a POST request, trying with adminAccess=true for Workflows first.
func (c Client) postWithAdminAccess(ctx context.Context, rt ResourceType, path string, data []byte) (*http.Response, error) {
	return c.requestWithAdminAccess(rt, func(opts rest.RequestOptions) (*http.Response, error) {
		return c.restClient.POST(ctx, path, bytes.NewReader(data), opts)
	})
}

// putWithAdminAccess performs a PUT request, trying with adminAccess=true for Workflows first.
func (c Client) putWithAdminAccess(ctx context.Context, rt ResourceType, path string, data []byte) (*http.Response, error) {
	return c.requestWithAdminAccess(rt, func(opts rest.RequestOptions) (*http.Response, error) {
		return c.restClient.PUT(ctx, path, bytes.NewReader(data), opts)
	})
}

// deleteWithAdminAccess performs a DELETE request, trying with adminAccess=true for Workflows first.
// It also skips retries on 404 responses.
func (c Client) deleteWithAdminAccess(ctx context.Context, rt ResourceType, path string) (*http.Response, error) {
	return c.requestWithAdminAccess(rt, func(opts rest.RequestOptions) (*http.Response, error) {
		deleteOpts := rest.RequestOptions{
			QueryParams: opts.QueryParams,
			CustomShouldRetryFunc: func(resp *http.Response) bool {
				if opts.CustomShouldRetryFunc == nil {
					return rest.RetryOnFailureExcept404(resp)
				}
				return opts.CustomShouldRetryFunc(resp) && rest.RetryOnFailureExcept404(resp)
			},
		}
		return c.restClient.DELETE(ctx, path, deleteOpts)
	})
}

func (c Client) requestWithAdminAccess(rt ResourceType, request func(opts rest.RequestOptions) (*http.Response, error)) (*http.Response, error) {
	if rt == Workflows {
		opts := rest.RequestOptions{
			QueryParams: url.Values{"adminAccess": []string{"true"}},
			CustomShouldRetryFunc: func(resp *http.Response) bool {
				return rest.RetryIfNotSuccess(resp) && (resp.StatusCode != http.StatusForbidden)
			},
		}
		resp, err := request(opts)
		if err != nil {
			return nil, err
		}
		if resp != nil && resp.StatusCode == http.StatusForbidden {
			resp.Body.Close()
			return request(rest.RequestOptions{})
		}
		return resp, nil
	}

	return request(rest.RequestOptions{})
}

func removeIDField(data *[]byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(*data, &m); err != nil {
		return err
	}
	delete(m, "id")
	var err error
	*data, err = json.Marshal(m)
	return err
}

func isAPIError(err error, target *api.APIError) bool {
	for err != nil {
		if apiErr, ok := err.(api.APIError); ok {
			*target = apiErr
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}
