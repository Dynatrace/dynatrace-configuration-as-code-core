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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const endpointPath = "platform/slo/v1/slos"

func NewClient(client *rest.Client) *Client {
	c := &Client{
		restClient: client,
	}
	return c
}

type Client struct {
	restClient *rest.Client
}

func (c *Client) List(ctx context.Context) (api.PagedListResponse, error) {
	var retVal api.PagedListResponse

	listPage := func(ctx context.Context, c *Client, pageKey string) (nextPage, api.ListResponse, error) {
		resp, err := c.restClient.GET(ctx, endpointPath, rest.RequestOptions{
			CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
			QueryParams:           url.Values{"page-key": {pageKey}}})
		if err != nil {
			return "", api.ListResponse{}, fmt.Errorf(errMsg, "list", err)
		}
		defer resp.Body.Close()

		return processListResponse(resp, unmarshallFromListResponse)
	}

	for haveNextPage, nextPageKey := true, ""; haveNextPage; {
		var listResponse api.ListResponse
		var err error

		nextPageKey, listResponse, err = listPage(ctx, c, nextPageKey)
		if err != nil {
			return nil, fmt.Errorf(errMsg, "list", err)
		}

		retVal = append(retVal, listResponse)
		haveNextPage = nextPageKey != ""
	}

	return retVal, nil

}

func (c *Client) Get(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf(errMsgWithId, "get", id, errors.New("argument \"id\" is empty"))
	}

	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "get", id, err)
	}

	resp, err := c.restClient.GET(ctx, path, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "get", id, err)
	}
	defer resp.Body.Close()

	return api.ProcessResponse(resp)
}

func (c *Client) Create(ctx context.Context, body json.RawMessage) (api.Response, error) {
	resp, err := c.restClient.POST(ctx, endpointPath, bytes.NewReader(body), rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsg, "create", err)
	}
	defer resp.Body.Close()

	return api.ProcessResponse(resp)
}

func (c *Client) Update(ctx context.Context, id string, body json.RawMessage) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf(errMsgWithId, "update", id, errors.New("argument \"id\" is empty"))
	}

	getResp, err := c.Get(ctx, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}

	version, err := getOptimisticLockingVersion(getResp)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}

	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}

	resp, err := c.restClient.PUT(ctx, path, bytes.NewReader(body), rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           url.Values{"optimistic-locking-version": []string{version}},
	})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}
	defer resp.Body.Close()

	return api.ProcessResponse(resp)
}

func (c *Client) Delete(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf(errMsgWithId, "delete", id, errors.New("argument \"id\" is empty"))
	}

	getResp, err := c.Get(ctx, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "delete", id, err)
	}

	version, err := getOptimisticLockingVersion(getResp)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "delete", id, err)
	}

	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "delete", id, err)
	}

	resp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           url.Values{"optimistic-locking-version": []string{version}},
	})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "delete", id, err)
	}
	defer resp.Body.Close()

	return api.ProcessResponse(resp)
}

const errMsg = "failed to %s slo resource: %w"
const errMsgWithId = "failed to %s slo resource with id %s: %w"

type nextPage = string

func getOptimisticLockingVersion(resp api.Response) (string, error) {
	var getStr struct {
		Version string `json:"version"`
	}

	err := json.Unmarshal(resp.Data, &getStr)
	if err != nil {
		return "", err
	}

	return getStr.Version, nil
}

func processListResponse(httpResponse *http.Response, transform func(json.RawMessage) (nextPage string, data [][]byte, e error)) (nextPage, api.ListResponse, error) {
	var err error

	if !rest.IsSuccess(httpResponse) {
		return "", api.ListResponse{}, api.NewAPIErrorFromResponse(httpResponse)
	}

	var body json.RawMessage
	if body, err = io.ReadAll(httpResponse.Body); err != nil {
		return "", api.ListResponse{}, api.NewAPIErrorFromResponse(httpResponse)
	}

	var data [][]byte
	var np nextPage
	if np, data, err = transform(body); err != nil {
		return "", api.ListResponse{}, api.NewAPIErrorFromResponse(httpResponse)
	}

	return np, api.NewListResponse(httpResponse, data), nil
}
func unmarshallFromListResponse(in json.RawMessage) (string, [][]byte, error) {
	var s struct {
		NextPage string            `json:"nextPageKey"`
		Data     []json.RawMessage `json:"slos"`
	}

	if err := json.Unmarshal(in, &s); err != nil {
		return "", nil, err
	}

	var data [][]byte
	for _, it := range s.Data {
		data = append(data, it)
	}

	return s.NextPage, data, nil
}
