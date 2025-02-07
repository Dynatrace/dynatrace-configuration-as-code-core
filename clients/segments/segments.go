// @license
// Copyright 2024 Dynatrace LLC
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

package segments

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/segments"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

type Response = api.Response

const errMsg = "failed to %s segments: %w"
const errMsgWithId = "failed to %s segments resource with id %s: %w"

func NewClient(client *rest.Client) *Client {
	c := &Client{
		client: segments.NewClient(client),
	}
	return c
}

// Client can be used to interact with the Automation API
type Client struct {
	client client
}

//go:generate mockgen -source segments.go -package=segments -destination=client_mock.go
type client interface {
	List(ctx context.Context, ro rest.RequestOptions) (*http.Response, error)
	Get(ctx context.Context, id string, ro rest.RequestOptions) (*http.Response, error)
	Create(ctx context.Context, data []byte, ro rest.RequestOptions) (*http.Response, error)
	Update(ctx context.Context, id string, data []byte, ro rest.RequestOptions) (*http.Response, error)
	Delete(ctx context.Context, id string, ro rest.RequestOptions) (*http.Response, error)
}

var _ client = (*segments.Client)(nil)

// List gets a complete set of available configs. The Data filed in response is normalized to json list of entries.
func (c Client) List(ctx context.Context) (Response, error) {
	resp, err := c.client.List(ctx, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return Response{}, fmt.Errorf(errMsg, "list", err)
	}

	apiResp, err := api.NewResponseFromHTTPResponse(resp)
	if err != nil {
		return apiResp, fmt.Errorf(errMsg, "list", err)
	}

	if apiResp.Data, err = modifyBody(apiResp.Data); err != nil {
		return apiResp, fmt.Errorf(errMsg, "list", err)
	}

	return apiResp, nil
}

// modifyBody transform received json response to contain just a JSON list of elements
func modifyBody(source []byte) ([]byte, error) {
	var transformed map[string]any
	if err := json.Unmarshal(source, &transformed); err != nil {
		return source, err
	}
	body, err := json.Marshal(transformed["filterSegments"])
	if err != nil {
		return source, err
	}
	return body, nil
}

// Get gets a complete configuration of segment with an ID
func (c Client) Get(ctx context.Context, id string) (Response, error) {
	resp, err := c.client.Get(ctx, id, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return Response{}, fmt.Errorf(errMsgWithId, "get", id, err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// GetAll gets a complete set of complete configuration for all available segments
func (c Client) GetAll(ctx context.Context) ([]Response, error) {
	listResp, err := c.List(ctx)
	if err != nil {
		return nil, err
	}

	var segments []struct {
		Uid string `json:"uid"`
	}
	if err = json.Unmarshal(listResp.Data, &segments); err != nil {
		return nil, fmt.Errorf(errMsg, "get all", err)
	}

	var result []Response
	for _, s := range segments {
		resp, err := c.Get(ctx, s.Uid)
		if err != nil {
			return nil, fmt.Errorf(errMsg, "get all", err)
		}
		result = append(result, resp)
	}

	return result, nil
}

func (c Client) Update(ctx context.Context, id string, data []byte) (Response, error) {
	existing, err := c.Get(ctx, id)
	if err != nil {
		return Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}

	var getResponse struct {
		Version int    `json:"version"`
		Owner   string `json:"owner"`
	}
	err = json.Unmarshal(existing.Data, &getResponse)
	if err != nil || getResponse.Version == 0 || getResponse.Owner == "" {
		return Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}

	// Adds owner if not set(they are mandatory from the API),
	// and always override uid with the uid that is getting updated
	data, err = addOwnerAndUIDIfNotSet(data, getResponse.Owner, id)
	if err != nil {
		return Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}

	updateResourceResp, err := c.client.Update(ctx, id, data, rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           map[string][]string{"optimistic-locking-version": {fmt.Sprint(getResponse.Version)}}})
	if err != nil {
		return Response{}, fmt.Errorf("failed to update segments resource with id %s and version %d: %w", id, getResponse.Version, err)
	}

	return api.NewResponseFromHTTPResponse(updateResourceResp)
}

// Create creates a new segment entry on the server
func (c Client) Create(ctx context.Context, data []byte) (Response, error) {
	resp, err := c.client.Create(ctx, data, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return Response{}, fmt.Errorf(errMsg, "create", err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// Delete removes configuration for segment with given ID from a server.
func (c Client) Delete(ctx context.Context, id string) (Response, error) {
	resp, err := c.client.Delete(ctx, id, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return Response{}, fmt.Errorf(errMsgWithId, "delete", id, err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func unmarshalRequest(payload []byte) (map[string]any, error) {
	var request map[string]any
	err := json.Unmarshal(payload, &request)
	if err != nil {
		return request, fmt.Errorf("failed to unmarshal request payload: %w", err)
	}
	return request, nil
}

func addOwnerAndUIDIfNotSet(payload []byte, owner string, uid string) ([]byte, error) {
	request, err := unmarshalRequest(payload)
	if err != nil {
		return nil, err
	}
	_, ok := request["owner"]
	if !ok {
		request["owner"] = owner
	}

	request["uid"] = uid

	return json.Marshal(request)
}
