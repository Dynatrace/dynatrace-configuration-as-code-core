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
	"context"
	"encoding/json"
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/openpipeline"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/go-logr/logr"
	"io"
	"slices"
)

const bodyReadErrMsg = "unable to read API response body"

type Response = api.Response

type ListResponse struct {
	Id       string `json:"id"`
	Editable bool   `json:"editable"`
}

type GetOptions struct {
	Editable bool
}

type GetAllOptions struct {
	Editable bool
}

type ListOptions struct {
	Editable bool
}

type UpdateOptions struct{}

func NewClient(client *rest.Client) *Client {
	c := &Client{
		client: openpipeline.NewClient(client),
	}

	return c
}

// Client can be used to interact with the Automation API
type Client struct {
	client *openpipeline.Client
}

func (c Client) Get(ctx context.Context, id string, _ GetOptions) (Response, error) {
	resp, err := c.client.Get(ctx, id, rest.RequestOptions{})
	if err != nil {
		return Response{}, fmt.Errorf("failed to get openpipeline resource of type id %q: %w", id, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
		return Response{}, api.NewAPIErrorFromResponseAndBody(resp, body)
	}

	if !rest.IsSuccess(resp) {
		return Response{}, api.NewAPIErrorFromResponseAndBody(resp, body)
	}

	return api.NewResponseFromHTTPResponseAndBody(resp, body), nil
}

func (c Client) List(ctx context.Context, options ListOptions) ([]ListResponse, error) {
	resp, err := c.client.List(ctx, rest.RequestOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list openpipeline resources: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
		return nil, api.NewAPIErrorFromResponseAndBody(resp, body)
	}

	var resources []ListResponse
	err = json.Unmarshal(body, &resources)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "failed to unmarshal json response")
		return nil, api.NewAPIErrorFromResponseAndBody(resp, body)
	}
	return slices.DeleteFunc(resources, func(response ListResponse) bool { return response.Editable != options.Editable }), nil
}

func (c Client) GetAll(ctx context.Context, options GetAllOptions) ([]Response, error) {
	listResp, err := c.List(ctx, ListOptions{Editable: options.Editable})
	if err != nil {
		return nil, err
	}

	var resources []Response
	for _, r := range listResp {
		if r.Editable != options.Editable {
			continue
		}
		rr, err := c.Get(ctx, r.Id, GetOptions{})
		if err != nil {
			return nil, err
		}
		resources = append(resources, rr)
	}
	return resources, nil
}

func (c Client) Update(ctx context.Context, id string, data []byte, _ UpdateOptions) (Response, error) {
	getResp, err := c.Get(ctx, id, GetOptions{})
	if err != nil {
		return Response{}, err
	}

	var m map[string]interface{}
	err = json.Unmarshal(getResp.Data, &m)
	if err != nil {
		return Response{}, err
	}

	var d map[string]interface{}
	err = json.Unmarshal(data, &d)
	if err != nil {
		return Response{}, err
	}

	d["version"] = m["version"]
	data, err = json.Marshal(d)
	if err != nil {
		return Response{}, fmt.Errorf("unable to marshal data: %w", err)
	}

	resp, err := c.client.Update(ctx, id, data, rest.RequestOptions{})
	if err != nil {
		return Response{}, fmt.Errorf("failed to list openpipeline resources: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
		return Response{}, api.NewAPIErrorFromResponseAndBody(resp, body)
	}

	if !rest.IsSuccess(resp) {
		return Response{}, api.NewAPIErrorFromResponseAndBody(resp, body)
	}

	return api.NewResponseFromHTTPResponseAndBody(resp, body), nil
}
