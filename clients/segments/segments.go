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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	segmentsResourcePath = "/platform/storage/filter-segments/v1/filter-segments"
	resource             = "segments"
)

var idValidationErr = api.ValidationError{Resource: resource, Field: "id", Reason: "is empty"}

// Client is used to interact with the Segments API.
type Client struct {
	restClient *rest.Client
}

// NewClient creates a new segments Client using the given rest.Client.
func NewClient(client *rest.Client) *Client {
	return &Client{restClient: client}
}

func (c Client) List(ctx context.Context) (api.Response, error) {
	resp, err := c.restClient.GET(ctx, segmentsResourcePath+":lean", rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           url.Values{"add-fields": []string{"EXTERNALID"}},
	})

	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	apiResp, err := api.NewResponseFromHTTPResponse(resp)
	if err != nil {
		return apiResp, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	apiResp.Data, err = modifyBody(apiResp.Data)
	if err != nil {
		return apiResp, api.RuntimeError{Resource: resource, Reason: "body transformation failed", Wrapped: err}
	}

	return apiResp, nil
}

// modifyBody transform received json response to contain just a JSON list of elements
func modifyBody(source []byte) ([]byte, error) {
	var transformed map[string]any
	if err := json.Unmarshal(source, &transformed); err != nil {
		return source, err
	}
	body, err := json.Marshal(transformed["filterSegments"])
	if err != nil {
		return source, err
	}
	return body, nil
}

func (c Client) Get(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr
	}

	path, err := url.JoinPath(segmentsResourcePath, id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.GET(ctx, path, rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           url.Values{"add-fields": []string{"INCLUDES", "VARIABLES", "EXTERNALID", "RESOURCECONTEXT"}},
	})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}
	return resp, nil
}

// Create creates a new segment.
func (c Client) Create(ctx context.Context, data []byte) (api.Response, error) {
	httpResp, err := c.restClient.POST(ctx, segmentsResourcePath, bytes.NewReader(data), rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPost, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPost, Wrapped: err}
	}
	return resp, nil
}

// Update updates an existing segment by ID.
func (c Client) Update(ctx context.Context, id string, data []byte) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr
	}

	existing, err := c.Get(ctx, id)
	if err != nil {
		return api.Response{}, err
	}

	var getResponse struct {
		Version int    `json:"version"`
		Owner   string `json:"owner"`
	}
	err = json.Unmarshal(existing.Data, &getResponse)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to unmarshal existing owner and version", Wrapped: err}
	}

	if getResponse.Version == 0 {
		return api.Response{}, api.ValidationError{Resource: resource, Field: "version", Reason: "is invalid"}
	}
	if getResponse.Owner == "" {
		return api.Response{}, api.ValidationError{Resource: resource, Field: "owner", Reason: "is empty"}
	}

	data, err = addOwnerAndUIDIfNotSet(data, getResponse.Owner, id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to add owner and UID", Wrapped: err}
	}

	path, err := url.JoinPath(segmentsResourcePath, id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.PUT(ctx, path, bytes.NewReader(data), rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           map[string][]string{"optimistic-locking-version": {strconv.Itoa(getResponse.Version)}},
	})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPut, Identifier: id, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodPut, Wrapped: err}
	}
	return resp, nil
}

func addOwnerAndUIDIfNotSet(payload []byte, owner string, uid string) ([]byte, error) {
	var request map[string]any
	if err := json.Unmarshal(payload, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request payload: %w", err)
	}
	if _, ok := request["owner"]; !ok {
		request["owner"] = owner
	}
	request["uid"] = uid
	newpayload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}
	return newpayload, nil
}

// Delete removes a given segment by ID.
func (c Client) Delete(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr
	}

	path, err := url.JoinPath(segmentsResourcePath, id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}
	return resp, nil
}

func (c Client) GetAll(ctx context.Context) ([]api.Response, error) {
	listResp, err := c.List(ctx)
	if err != nil {
		return nil, err
	}

	var segments []struct {
		Uid string `json:"uid"`
	}
	if err = json.Unmarshal(listResp.Data, &segments); err != nil {
		return nil, api.RuntimeError{Resource: resource, Reason: "unmarshalling failed", Wrapped: err}
	}

	var result []api.Response
	for _, s := range segments {
		resp, err := c.Get(ctx, s.Uid)
		if err != nil {
			return nil, err
		}
		result = append(result, resp)
	}

	return result, nil
}
