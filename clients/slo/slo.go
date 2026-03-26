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
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	sloResourcePath = "/platform/slo/v1/slos"
	resource        = "slo"
)

var idValidationErr = api.ValidationError{Resource: resource, Field: "id", Reason: "is empty"}

// Client is used to interact with the SLO API.
type Client struct {
	restClient *rest.Client
}

// NewClient creates a new SLO Client using the given rest.Client.
func NewClient(client *rest.Client) *Client {
	return &Client{restClient: client}
}

// List returns all SLO objects.
func (c Client) List(ctx context.Context) (api.PagedListResponse, error) {
	var pagedListResponse api.PagedListResponse

	nextPageKey := ""
	for {
		var listResponse api.ListResponse
		var err error

		nextPageKey, listResponse, err = c.listPage(ctx, nextPageKey)
		if err != nil {
			return nil, err
		}

		pagedListResponse = append(pagedListResponse, listResponse)
		if nextPageKey == "" {
			break
		}
	}

	return pagedListResponse, nil
}

func (c Client) listPage(ctx context.Context, pageKey string) (string, api.ListResponse, error) {
	ro := rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests}
	if pageKey != "" {
		ro.QueryParams = url.Values{"page-key": {pageKey}}
	}

	resp, err := c.restClient.GET(ctx, sloResourcePath, ro)
	if err != nil {
		return "", api.ListResponse{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	return processListResponse(resp)
}

func processListResponse(httpResponse *http.Response) (string, api.ListResponse, error) {
	resp, err := api.NewResponseFromHTTPResponse(httpResponse)
	if err != nil {
		return "", api.ListResponse{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	var sloResponse struct {
		NextPage string            `json:"nextPageKey"`
		SLOs     []json.RawMessage `json:"slos"`
	}

	if err := json.Unmarshal(resp.Data, &sloResponse); err != nil {
		return "", api.ListResponse{}, api.RuntimeError{Resource: resource, Reason: "unmarshalling failed", Wrapped: err}
	}

	var objects [][]byte
	for _, it := range sloResponse.SLOs {
		objects = append(objects, it)
	}

	return sloResponse.NextPage,
		api.ListResponse{
			Response: api.Response{
				StatusCode: httpResponse.StatusCode,
				Header:     httpResponse.Header,
				Data:       nil,
				Request:    api.NewRequestInfoFromRequest(httpResponse.Request),
			},
			Objects: objects,
		},
		nil
}

// Get returns one specific SLO object by ID.
func (c Client) Get(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr
	}

	path, err := url.JoinPath(sloResourcePath, id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.GET(ctx, path, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}
	return resp, nil
}

// Create creates a new SLO.
func (c Client) Create(ctx context.Context, data []byte) (api.Response, error) {
	httpResp, err := c.restClient.POST(ctx, sloResourcePath, bytes.NewReader(data), rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPost, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPost, Wrapped: err}
	}
	return resp, nil
}

// Update updates an existing SLO by ID.
func (c Client) Update(ctx context.Context, id string, data []byte) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr
	}

	existing, err := c.Get(ctx, id)
	if err != nil {
		return api.Response{}, err
	}

	version, err := getOptimisticLockingVersion(existing)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to retrieve optimistic locking version", Wrapped: err}
	}

	path, err := url.JoinPath(sloResourcePath, id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.PUT(ctx, path, bytes.NewReader(data), rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           url.Values{"optimistic-locking-version": {version}},
	})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodPut, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodPut, Wrapped: err}
	}
	return resp, nil
}

// Delete removes a given SLO by ID.
func (c Client) Delete(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr
	}

	existing, err := c.Get(ctx, id)
	if err != nil {
		return api.Response{}, err
	}

	version, err := getOptimisticLockingVersion(existing)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to retrieve optimistic locking version", Wrapped: err}
	}

	path, err := url.JoinPath(sloResourcePath, id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           url.Values{"optimistic-locking-version": {version}},
	})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}
	return resp, nil
}

func getOptimisticLockingVersion(resp api.Response) (string, error) {
	var body struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(resp.Data, &body); err != nil {
		return "", err
	}

	return body.Version, nil
}
