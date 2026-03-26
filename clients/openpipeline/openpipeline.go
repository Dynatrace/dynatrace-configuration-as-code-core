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
	"net/http"
	"net/url"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	maxUpdateAttempts        = 10
	openPipelineResourcePath = "/platform/openpipeline/v1/configurations"
	resource                 = "openpipeline"
)

var idValidationErr = api.ValidationError{Resource: resource, Field: "id", Reason: "is empty"}

// Client is used to interact with the OpenPipeline API.
type Client struct {
	restClient *rest.Client
}

// NewClient creates a new OpenPipeline Client using the given rest.Client.
func NewClient(client *rest.Client) *Client {
	return &Client{restClient: client}
}

// Get returns one specific OpenPipeline configuration by ID.
func (c Client) Get(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr
	}

	path, err := url.JoinPath(openPipelineResourcePath, id)
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

// List returns all OpenPipeline configurations.
func (c Client) List(ctx context.Context) (api.Response, error) {
	httpResp, err := c.restClient.GET(ctx, openPipelineResourcePath, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}
	return resp, nil
}

// GetAll returns all OpenPipeline configurations with their full details.
func (c Client) GetAll(ctx context.Context) ([]api.Response, error) {
	listResp, err := c.List(ctx)
	if err != nil {
		return nil, err
	}

	var configurations []struct {
		Id string `json:"id"`
	}
	if err = json.Unmarshal(listResp.Data, &configurations); err != nil {
		return nil, api.RuntimeError{Resource: resource, Reason: "unmarshalling failed", Wrapped: err}
	}

	var resources []api.Response
	for _, r := range configurations {
		rr, err := c.Get(ctx, r.Id)
		if err != nil {
			return nil, err
		}
		resources = append(resources, rr)
	}
	return resources, nil
}

// Update updates an existing OpenPipeline configuration by ID.
// It retries on conflict errors up to a maximum number of attempts.
func (c Client) Update(ctx context.Context, id string, payload []byte) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr
	}

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
	if err = json.Unmarshal(getResp.Data, &remoteJson); err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to unmarshal GET response", Wrapped: err}
	}

	var localJson map[string]any
	if err = json.Unmarshal(payload, &localJson); err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to unmarshal request payload", Wrapped: err}
	}

	localJson["version"] = remoteJson["version"]
	localJson["updateToken"] = remoteJson["updateToken"]
	mergedPayload, err := json.Marshal(localJson)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to marshal payload", Wrapped: err}
	}

	path, err := url.JoinPath(openPipelineResourcePath, id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.PUT(ctx, path, bytes.NewReader(mergedPayload), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodPut, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodPut, Wrapped: err}
	}
	return resp, nil
}
