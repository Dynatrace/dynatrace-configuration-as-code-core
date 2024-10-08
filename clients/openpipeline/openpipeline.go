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
)

type Response = api.Response

type ListResponse struct {
	Id       string `json:"id"`
	Editable bool   `json:"editable"`
}

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

func (c Client) Get(ctx context.Context, id string) (*Response, error) {
	return api.AsResponseOrError(c.client.Get(ctx, id, rest.RequestOptions{}))
}

func (c Client) List(ctx context.Context) ([]ListResponse, error) {
	resp, err := api.AsResponseOrError(c.client.List(ctx, rest.RequestOptions{}))
	if err != nil {
		return nil, err
	}

	var resources []ListResponse
	err = json.Unmarshal(resp.Data, &resources)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "failed to unmarshal json response")
		return nil, fmt.Errorf("failed to unmarshal json response: %w", err)
	}
	return resources, nil
}

func (c Client) GetAll(ctx context.Context) ([]Response, error) {
	listResp, err := c.List(ctx)
	if err != nil {
		return nil, err
	}

	var resources []Response
	for _, r := range listResp {
		rr, err := c.Get(ctx, r.Id)
		if err != nil {
			return nil, err
		}
		resources = append(resources, *rr)
	}
	return resources, nil
}

func (c Client) Update(ctx context.Context, id string, data []byte) (*Response, error) {
	getResp, err := c.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	err = json.Unmarshal(getResp.Data, &m)
	if err != nil {
		return nil, err
	}

	var d map[string]interface{}
	err = json.Unmarshal(data, &d)
	if err != nil {
		return nil, err
	}

	d["version"] = m["version"]
	data, err = json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal data: %w", err)
	}

	return api.AsResponseOrError(c.client.Update(ctx, id, data, rest.RequestOptions{}))
}
