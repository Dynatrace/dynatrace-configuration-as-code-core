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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const endpointPath = "platform/storage/filter-segments/v1/filter-segments"

const errMsg = "failed to %s segments: %w"
const errMsgWithId = "failed to %s segments resource with id %s: %w"

func NewClient(client *rest.Client) *Client {
	return &Client{restClient: client}
}

type Client struct {
	restClient *rest.Client
}

func (c Client) List(ctx context.Context) (api.Response, error) {
	path := endpointPath + ":lean" // minimal set of information is enough

	ro := rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           url.Values{"add-fields": []string{"EXTERNALID"}},
	}

	resp, err := c.restClient.GET(ctx, path, ro)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsg, "list", err)
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

func (c Client) Get(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf(errMsgWithId, "get", id, errors.New(`argument "id" is empty`))
	}

	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "get", id, err)
	}

	ro := rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           url.Values{"add-fields": []string{"INCLUDES", "VARIABLES", "EXTERNALID", "RESOURCECONTEXT"}},
	}

	resp, err := c.restClient.GET(ctx, path, ro)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "get", id, err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func (c Client) Create(ctx context.Context, body []byte) (api.Response, error) {
	resp, err := c.restClient.POST(ctx, endpointPath, bytes.NewReader(body), rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsg, "create", err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func (c Client) Update(ctx context.Context, id string, body []byte) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf(errMsgWithId, "update", id, errors.New(`argument "id" is empty`))
	}

	existing, err := c.Get(ctx, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}

	var getResponse struct {
		Version int    `json:"version"`
		Owner   string `json:"owner"`
	}
	err = json.Unmarshal(existing.Data, &getResponse)
	if err != nil || getResponse.Version == 0 || getResponse.Owner == "" {
		return api.Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}

	// Adds owner if not set(they are mandatory from the API),
	// and always override uid with the uid that is getting updated
	body, err = addOwnerAndUIDIfNotSet(body, getResponse.Owner, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}

	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "update", id, err)
	}

	resp, err := c.restClient.PUT(ctx, path, bytes.NewReader(body), rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           map[string][]string{"optimistic-locking-version": {fmt.Sprint(getResponse.Version)}}})
	if err != nil {
		return api.Response{}, fmt.Errorf("failed to update segments resource with id %s and version %d: %w", id, getResponse.Version, err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func (c Client) Delete(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf(errMsgWithId, "delete", id, errors.New(`argument "id" is empty`))
	}

	path, err := url.JoinPath(endpointPath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "delete", id, err)
	}

	resp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithId, "delete", id, err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func (c Client) GetAll(ctx context.Context) ([]api.Response, error) {
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

	var result []api.Response
	for _, s := range segments {
		resp, err := c.Get(ctx, s.Uid)
		if err != nil {
			return nil, fmt.Errorf(errMsg, "get all", err)
		}
		result = append(result, resp)
	}

	return result, nil
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
