// @license
// Copyright 2026 Dynatrace LLC
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

package directshares

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
	directSharesResourcePath = "/platform/document/v1/direct-shares"
	resource                 = "direct-shares"
)

var idValidationErr = api.ValidationError{Resource: resource, Field: "id", Reason: "is empty"}

type Client struct {
	restClient *rest.Client
}

// NewClient creates a new direct shares Client using the given rest.Client.
func NewClient(client *rest.Client) *Client {
	return &Client{restClient: client}
}

// List returns all direct share objects.
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
	ro := rest.RequestOptions{}
	if pageKey != "" {
		ro.QueryParams = url.Values{"page-key": {pageKey}}
	}

	resp, err := c.restClient.GET(ctx, directSharesResourcePath, ro)
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

	var directSharesResponse struct {
		NextPage     string            `json:"nextPageKey"`
		DirectShares []json.RawMessage `json:"directShares"`
	}

	if err := json.Unmarshal(resp.Data, &directSharesResponse); err != nil {
		return "", api.ListResponse{}, api.RuntimeError{Resource: resource, Reason: "unmarshalling failed", Wrapped: err}
	}

	var objects [][]byte
	for _, it := range directSharesResponse.DirectShares {
		objects = append(objects, it)
	}

	return directSharesResponse.NextPage,
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

// Get returns one specific direct share object by ID.
func (c Client) Get(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr
	}

	path, err := url.JoinPath(directSharesResourcePath, id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}
	return resp, nil
}

// GetRecipients returns the recipients of a specific direct share object by ID.
func (c Client) GetRecipients(ctx context.Context, id string) (api.PagedListResponse, error) {
	if id == "" {
		return nil, idValidationErr
	}

	var pagedListResponse api.PagedListResponse

	nextPageKey := ""
	for {
		var listResponse api.ListResponse
		var err error

		nextPageKey, listResponse, err = c.listRecipientsPage(ctx, id, nextPageKey)
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

func (c Client) listRecipientsPage(ctx context.Context, id string, pageKey string) (string, api.ListResponse, error) {
	ro := rest.RequestOptions{}
	if pageKey != "" {
		ro.QueryParams = url.Values{"page-key": {pageKey}}
	}

	path, err := url.JoinPath(directSharesResourcePath, id, "recipients")
	if err != nil {
		return "", api.ListResponse{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.GET(ctx, path, ro)
	if err != nil {
		return "", api.ListResponse{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}

	return processRecipientsListResponse(httpResp)
}

func processRecipientsListResponse(httpResponse *http.Response) (string, api.ListResponse, error) {
	resp, err := api.NewResponseFromHTTPResponse(httpResponse)
	if err != nil {
		return "", api.ListResponse{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	var recipientsResponse struct {
		NextPage   string            `json:"nextPageKey"`
		Recipients []json.RawMessage `json:"recipients"`
	}

	if err := json.Unmarshal(resp.Data, &recipientsResponse); err != nil {
		return "", api.ListResponse{}, api.RuntimeError{Resource: resource, Reason: "unmarshalling failed", Wrapped: err}
	}

	var objects [][]byte
	for _, it := range recipientsResponse.Recipients {
		objects = append(objects, it)
	}

	return recipientsResponse.NextPage,
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

// AddRecipients adds recipients to a specific direct share.
func (c Client) AddRecipients(ctx context.Context, id string, data []byte) error {
	if id == "" {
		return idValidationErr
	}

	path, err := url.JoinPath(directSharesResourcePath, id, "recipients", "add")
	if err != nil {
		return api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.POST(ctx, path, bytes.NewReader(data), rest.RequestOptions{})
	if err != nil {
		return api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodPost, Wrapped: err}
	}

	_, err = api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodPost, Wrapped: err}
	}
	return nil
}

// RemoveRecipients removes recipients from a specific direct share.
func (c Client) RemoveRecipients(ctx context.Context, id string, data []byte) error {
	if id == "" {
		return idValidationErr
	}

	path, err := url.JoinPath(directSharesResourcePath, id, "recipients", "remove")
	if err != nil {
		return api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.POST(ctx, path, bytes.NewReader(data), rest.RequestOptions{})
	if err != nil {
		return api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodPost, Wrapped: err}
	}

	_, err = api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodPost, Wrapped: err}
	}
	return nil
}

// Create creates a document direct share.
func (c Client) Create(ctx context.Context, data []byte) (api.Response, error) {
	httpResp, err := c.restClient.POST(ctx, directSharesResourcePath, bytes.NewReader(data), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPost, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPost, Wrapped: err}
	}
	return resp, nil
}

// Delete removes a given document direct share by ID.
func (c Client) Delete(ctx context.Context, id string) (err error) {
	if id == "" {
		return idValidationErr
	}

	path, err := url.JoinPath(directSharesResourcePath, id)
	if err != nil {
		return api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}

	_, err = api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}
	return nil
}
