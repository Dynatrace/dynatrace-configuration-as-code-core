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

package credentialvault

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

const (
	resourcePath      = "/platform/credential-vault/v1/entries"
	urlCreationErrMsg = "failed to construct URL"
	resource          = "credential-vault"
	pageSize          = 500
)

var (
	errEntryIDEmpty = api.ValidationError{Resource: resource, Field: "ID", Reason: "is empty"}
)

// Client is used to interact with the Credential Vault API.
type Client struct {
	restClient *rest.Client
}

// NewClient creates a new credential vault Client using the given rest.Client.
func NewClient(client *rest.Client) *Client {
	return &Client{restClient: client}
}

// List returns all credential vault entries.
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
	var ro rest.RequestOptions
	if pageKey != "" {
		ro.QueryParams = url.Values{"page-key": {pageKey}}
	} else {
		ro.QueryParams = url.Values{"page-size": {strconv.Itoa(pageSize)}}
	}

	httpResponse, err := c.restClient.GET(ctx, resourcePath, ro)
	if err != nil {
		return "", api.ListResponse{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResponse)
	if err != nil {
		return "", api.ListResponse{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	var listResp struct {
		NextPage string            `json:"nextPageKey"`
		Items    []json.RawMessage `json:"credentials"`
	}

	if err := json.Unmarshal(resp.Data, &listResp); err != nil {
		return "", api.ListResponse{}, api.RuntimeError{Resource: resource, Reason: "unmarshalling failed", Wrapped: err}
	}

	objects := make([][]byte, 0, len(listResp.Items))
	for _, it := range listResp.Items {
		objects = append(objects, it)
	}

	return listResp.NextPage,
		api.ListResponse{
			StatusCode: httpResponse.StatusCode,
			Header:     httpResponse.Header,
			Data:       nil,
			Request:    api.NewRequestInfoFromRequest(httpResponse.Request),
			Objects:    objects,
		},
		nil
}

// Get returns the credential vault entry for a given ID.
func (c Client) Get(ctx context.Context, entryID string) (api.Response, error) {
	if entryID == "" {
		return api.Response{}, errEntryIDEmpty
	}

	path, err := url.JoinPath(resourcePath, entryID)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: entryID, Reason: urlCreationErrMsg, Wrapped: err}
	}

	httpResp, err := c.restClient.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: entryID, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: entryID, Operation: http.MethodGet, Wrapped: err}
	}
	return resp, nil
}

// Create creates a new credential vault entry.
func (c Client) Create(ctx context.Context, data []byte) (api.Response, error) {
	httpResp, err := c.restClient.POST(ctx, resourcePath, bytes.NewReader(data), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPost, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPost, Wrapped: err}
	}
	return resp, nil
}

// Update updates an existing credential vault entry.
func (c Client) Update(ctx context.Context, entryID string, data []byte) (api.Response, error) {
	if entryID == "" {
		return api.Response{}, errEntryIDEmpty
	}

	path, err := url.JoinPath(resourcePath, entryID)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: entryID, Reason: urlCreationErrMsg, Wrapped: err}
	}

	httpResp, err := c.restClient.PUT(ctx, path, bytes.NewReader(data), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: entryID, Operation: http.MethodPut, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: entryID, Operation: http.MethodPut, Wrapped: err}
	}
	return resp, nil
}

// Delete deletes a specific credential vault entry.
func (c Client) Delete(ctx context.Context, entryID string) (api.Response, error) {
	if entryID == "" {
		return api.Response{}, errEntryIDEmpty
	}

	path, err := url.JoinPath(resourcePath, entryID)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: entryID, Reason: urlCreationErrMsg, Wrapped: err}
	}

	httpResp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: entryID, Operation: http.MethodDelete, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: entryID, Operation: http.MethodDelete, Wrapped: err}
	}
	return resp, nil
}
