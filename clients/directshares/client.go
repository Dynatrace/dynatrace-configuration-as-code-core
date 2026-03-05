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
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	directSharesResourcePath = "/platform/document/v1/direct-shares"

	errMsg = "failed to %s document direct share: %w"

	errMsgWithID = "failed to %s document direct share with ID %s: %w"

	getOperation              = "get"
	listOperation             = "list"
	createOperation           = "create"
	deleteOperation           = "delete"
	getRecipientsOperation    = "get recipients"
	addRecipientsOperation    = "add recipients"
	removeRecipientsOperation = "remove recipients"
)

var ErrIDEmpty = fmt.Errorf("id must be non-empty")

// ListResponse is a list of API Responses
type ListResponse struct {
	api.Response
	Responses []api.Response
}

type Client struct {
	client *rest.Client
}

func NewClient(client *rest.Client) *Client {
	return &Client{client: client}
}

// List returns all direct share objects.
func (c Client) List(ctx context.Context) (api.PagedListResponse, error) {
	var retVal api.PagedListResponse

	for hasNextPage, nextPageKey := true, ""; hasNextPage; {
		var listResponse api.ListResponse
		var err error

		nextPageKey, listResponse, err = c.listPage(ctx, nextPageKey)
		if err != nil {
			return nil, fmt.Errorf(errMsg, listOperation, err)
		}

		retVal = append(retVal, listResponse)
		hasNextPage = nextPageKey != ""
	}

	return retVal, nil
}

func (c Client) listPage(ctx context.Context, pageKey string) (string, api.ListResponse, error) {
	ro := rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests}
	if pageKey != "" {
		ro.QueryParams = url.Values{"page-key": {pageKey}}
	}

	resp, err := c.client.GET(ctx, directSharesResourcePath, ro)
	if err != nil {
		return "", api.ListResponse{}, fmt.Errorf(errMsg, listOperation, err)
	}

	return processListResponse(resp)
}

func processListResponse(httpResponse *http.Response) (nextPageKey string, resp api.ListResponse, err error) {
	defer httpResponse.Body.Close()

	var body json.RawMessage
	if body, err = io.ReadAll(httpResponse.Body); err != nil {
		return "", api.ListResponse{}, api.NewAPIErrorFromResponse(httpResponse)
	}

	if !rest.IsSuccess(httpResponse) {
		return "", api.ListResponse{}, api.NewAPIErrorFromResponseAndBody(httpResponse, body)
	}

	var s struct {
		NextPage string            `json:"nextPageKey"`
		Data     []json.RawMessage `json:"directShares"`
	}

	if err := json.Unmarshal(body, &s); err != nil {
		return "", api.ListResponse{}, api.NewAPIErrorFromResponseAndBody(httpResponse, body)
	}

	var data [][]byte
	for _, it := range s.Data {
		data = append(data, it)
	}

	resp = api.ListResponse{
		Response: api.Response{
			StatusCode: httpResponse.StatusCode,
			Header:     httpResponse.Header,
			Data:       nil,
			Request:    api.NewRequestInfoFromRequest(httpResponse.Request),
		},
		Objects: data,
	}

	return s.NextPage, resp, nil
}

// Get returns one specific direct share object by ID.
func (c Client) Get(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf(errMsg, getOperation, ErrIDEmpty)
	}

	path, err := url.JoinPath(directSharesResourcePath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, getOperation, id, err)
	}

	httpResp, err := c.client.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, getOperation, id, err)
	}

	return api.NewResponseFromHTTPResponse(httpResp)
}

// GetRecipients returns the recipients of a specific direct share object by ID.
func (c Client) GetRecipients(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf(errMsg, getRecipientsOperation, ErrIDEmpty)
	}

	path, err := url.JoinPath(directSharesResourcePath, id, "recipients")
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, getRecipientsOperation, id, err)
	}

	httpResp, err := c.client.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, getRecipientsOperation, id, err)
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, getRecipientsOperation, id, err)
	}

	return resp, nil
}

func (c Client) AddRecipients(ctx context.Context, id string, data []byte) error {
	if id == "" {
		return fmt.Errorf(errMsg, addRecipientsOperation, ErrIDEmpty)
	}

	path, err := url.JoinPath(directSharesResourcePath, id, "recipients", "add")
	if err != nil {
		return fmt.Errorf(errMsgWithID, addRecipientsOperation, id, err)
	}

	httpResp, err := c.client.POST(ctx, path, bytes.NewReader(data), rest.RequestOptions{})
	if err != nil {
		return fmt.Errorf(errMsgWithID, addRecipientsOperation, id, err)
	}

	_, err = api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return fmt.Errorf(errMsgWithID, addRecipientsOperation, id, err)
	}
	return nil
}

func (c Client) RemoveRecipients(ctx context.Context, id string, data []byte) error {
	if id == "" {
		return fmt.Errorf(errMsg, removeRecipientsOperation, ErrIDEmpty)
	}

	path, err := url.JoinPath(directSharesResourcePath, id, "recipients", "remove")
	if err != nil {
		return fmt.Errorf(errMsgWithID, removeRecipientsOperation, id, err)
	}

	httpResp, err := c.client.POST(ctx, path, bytes.NewReader(data), rest.RequestOptions{})
	if err != nil {
		return fmt.Errorf(errMsgWithID, removeRecipientsOperation, id, err)
	}

	_, err = api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return fmt.Errorf(errMsgWithID, removeRecipientsOperation, id, err)
	}
	return nil
}

// Create creates a given document object
func (c Client) Create(ctx context.Context, data []byte) (api.Response, error) {
	httpResp, err := c.client.POST(ctx, directSharesResourcePath, bytes.NewReader(data), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsg, createOperation, err)
	}
	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsg, createOperation, err)
	}

	return resp, nil
}

// Delete removes a given automation object by ID
func (c Client) Delete(ctx context.Context, id string) (err error) {
	if id == "" {
		return fmt.Errorf(errMsg, deleteOperation, ErrIDEmpty)
	}

	path, err := url.JoinPath(directSharesResourcePath, id)
	if err != nil {
		return fmt.Errorf(errMsgWithID, deleteOperation, id, err)
	}

	httpResp, err := c.client.DELETE(ctx, path, rest.RequestOptions{})
	if err != nil {
		return fmt.Errorf(errMsgWithID, deleteOperation, id, err)
	}
	_, err = api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return fmt.Errorf(errMsgWithID, deleteOperation, id, err)
	}

	return nil
}
