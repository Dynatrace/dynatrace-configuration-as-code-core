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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	httpClient "github.com/dynatrace/dynatrace-configuration-as-code-core/clients/slo/internal/client"
)

func NewClient(client *rest.Client) *Client {
	c := &Client{
		client: httpClient.NewClient(client),
	}
	return c
}

type Client struct {
	client client
}

//go:generate mockgen -source slo.go -package=slo -destination=client_mock.go
type client interface {
	List(ctx context.Context, ro rest.RequestOptions) (*http.Response, error)
	Get(ctx context.Context, id string, ro rest.RequestOptions) (*http.Response, error)
	Create(ctx context.Context, data []byte, ro rest.RequestOptions) (*http.Response, error)
	Update(ctx context.Context, id string, optimisticLockingVersion string, data []byte, ro rest.RequestOptions) (*http.Response, error)
	Delete(ctx context.Context, id string, ro rest.RequestOptions) (*http.Response, error)
}

var _ client = (*httpClient.Client)(nil)

const errMsgWithId = "failed to %s slo resource with id %s: %w"

// List gets a complete set of complete configuration for all available SLOs
func (c Client) List(ctx context.Context) (api.PagedListResponse, error) {
	var retVal api.PagedListResponse

	for haveNextPage, nextPageKey := true, ""; haveNextPage; {
		response, err := c.client.List(ctx, rest.RequestOptions{
			CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
			QueryParams:           url.Values{"page-key": {nextPageKey}}})
		defer closeBody(response)
		if err != nil {
			return nil, err
		}

		var listResponse api.ListResponse
		nextPageKey, listResponse, err = processListResponse(response, unmarshallFromListResponse)
		if err != nil {
			return nil, err
		}

		retVal = append(retVal, listResponse)
		haveNextPage = nextPageKey != ""
	}

	return retVal, nil
}

// Get gets a complete configuration of SLO with an ID
func (c Client) Get(ctx context.Context, id string) (api.Response, error) {
	resp, err := c.client.Get(ctx, id, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	defer closeBody(resp)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "get", id, err)
	}
	return api.ProcessResponse(resp, nil)
}

func (c Client) Create(ctx context.Context, body json.RawMessage) (api.Response, error) {
	resp, err := c.client.Create(ctx, body, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	defer closeBody(resp)
	if err != nil {
		return api.Response{}, err // error message from the client is descriptive enough
	}
	return api.ProcessResponse(resp, nil)
}

func (c Client) Update(ctx context.Context, id string, body json.RawMessage) (api.Response, error) {
	getResp, err := c.Get(ctx, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}

	version, err := getVersion(getResp)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}

	resp, err := c.client.Update(ctx, id, version, body, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	defer closeBody(resp)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}
	return api.ProcessResponse(resp, nil)
}

func getVersion(resp api.Response) (string, error) {
	var getStr struct {
		Version string `json:"version"`
	}

	err := json.Unmarshal(resp.Data, &getStr)
	if err != nil {
		return "", err
	}

	return getStr.Version, nil
}

// Delete removes configuration for SLO with given ID from a server.
func (c Client) Delete(ctx context.Context, id string) (api.Response, error) {
	resp, err := c.client.Delete(ctx, id, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	closeBody(resp)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "delete", id, err)
	}

	if !rest.IsSuccess(resp) {
		return api.Response{}, api.NewAPIErrorFromResponse(resp)
	}
	return api.NewResponseFromHTTPResponseAndBody(resp, nil), nil
}

func closeBody(httpResponse *http.Response) {
	if httpResponse != nil && httpResponse.Body != nil {
		_ = httpResponse.Body.Close()
	}
}

type nextPage = string

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
