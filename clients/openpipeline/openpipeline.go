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
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-logr/logr"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/openpipeline"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const bodyReadErrMsg = "unable to read API response body"
const maxUpdateAttempts = 10

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

func (c Client) Get(ctx context.Context, id string) (Response, error) {
	resp, err := c.client.Get(ctx, id, rest.RequestOptions{})
	if err != nil {
		return Response{}, fmt.Errorf("failed to get openpipeline resource of type id %q: %w", id, err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func (c Client) List(ctx context.Context) ([]ListResponse, error) {
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
		resources = append(resources, rr)
	}
	return resources, nil
}

func (c Client) Update(ctx context.Context, id string, payload []byte) (Response, error) {
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

func (c Client) update(ctx context.Context, id string, payload []byte) (Response, error) {
	getResp, err := c.Get(ctx, id)
	if err != nil {
		return Response{}, err
	}

	var remoteJson map[string]any
	err = json.Unmarshal(getResp.Data, &remoteJson)
	if err != nil {
		return Response{}, err
	}

	var localJson map[string]any
	err = json.Unmarshal(payload, &localJson)
	if err != nil {
		return Response{}, err
	}

	localJson["version"] = remoteJson["version"]
	localJson["updateToken"] = remoteJson["updateToken"]
	payload, err = json.Marshal(localJson)
	if err != nil {
		return Response{}, fmt.Errorf("unable to marshal payload: %w", err)
	}

	resp, err := c.client.Update(ctx, id, payload, rest.RequestOptions{})
	if err != nil {
		return Response{}, fmt.Errorf("failed to list openpipeline resources: %w", err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}
