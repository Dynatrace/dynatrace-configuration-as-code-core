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

package openpipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	maxUpdateAttempts        = 10
	openPipelineResourcePath = "/platform/openpipeline/v1/configurations"

	errMsg       = "failed to %s openpipeline resource: %w"
	errMsgWithId = "failed to %s openpipeline resource with id %s: %w"

	getOperation    = "get"
	updateOperation = "update"
	listOperation   = "list"
)

var ErrEmptyID = errors.New("id must be non-empty")

type ListResponse struct {
	Id       string `json:"id"`
	Editable bool   `json:"editable"`
}

func NewClient(client *rest.Client) *Client {
	c := &Client{
		restClient: client,
	}

	return c
}

// Client can be used to interact with the Automation API
type Client struct {
	restClient *rest.Client
}

func (c Client) Get(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf(errMsg, getOperation, ErrEmptyID)
	}

	path, err := url.JoinPath(openPipelineResourcePath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, getOperation, id, err)
	}

	resp, err := c.restClient.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, getOperation, id, err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func (c Client) List(ctx context.Context) ([]ListResponse, error) {
	httpResp, err := c.restClient.GET(ctx, openPipelineResourcePath, rest.RequestOptions{})
	if err != nil {
		return nil, fmt.Errorf(errMsg, listOperation, err)
	}
	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return nil, fmt.Errorf(errMsg, listOperation, err)
	}

	var resources []ListResponse
	err = json.Unmarshal(resp.Data, &resources)
	if err != nil {
		return nil, fmt.Errorf(errMsg, listOperation, err)
	}
	return resources, nil
}

func (c Client) GetAll(ctx context.Context) ([]api.Response, error) {
	listResp, err := c.List(ctx)
	if err != nil {
		return nil, err
	}

	var resources []api.Response
	for _, r := range listResp {
		rr, err := c.Get(ctx, r.Id)
		if err != nil {
			return nil, err
		}
		resources = append(resources, rr)
	}
	return resources, nil
}

func (c Client) Update(ctx context.Context, id string, payload []byte) (api.Response, error) {
	var resp api.Response
	var err error

	for i := 0; i < maxUpdateAttempts; i++ {
		resp, err = c.update(ctx, id, payload)
		if err == nil {
			return resp, nil
		}

		var apiErr api.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusConflict {
			return resp, err
		}
	}
	return resp, err
}

func (c Client) update(ctx context.Context, id string, payload []byte) (api.Response, error) {
	getResp, err := c.Get(ctx, id)
	if err != nil {
		return api.Response{}, err
	}

	var remoteJson map[string]any
	err = json.Unmarshal(getResp.Data, &remoteJson)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, updateOperation, id, fmt.Errorf("unable to unmarshal GET response: %w", err))
	}

	var localJson map[string]any
	err = json.Unmarshal(payload, &localJson)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, updateOperation, id, fmt.Errorf("unable to unmarshal request payload: %w", err))
	}

	localJson["version"] = remoteJson["version"]
	localJson["updateToken"] = remoteJson["updateToken"]
	payload, err = json.Marshal(localJson)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, updateOperation, id, fmt.Errorf("unable to marshal payload: %w", err))
	}

	path, err := url.JoinPath(openPipelineResourcePath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, updateOperation, id, err)
	}

	resp, err := c.restClient.PUT(ctx, path, bytes.NewReader(payload), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, updateOperation, id, err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}
