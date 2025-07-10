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
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const endpointPath = "platform/storage/filter-segments/v1/filter-segments"
const resource = "segments"

var basePath = url.URL{Path: endpointPath}

func NewClient(client *rest.Client) *Client {
	return &Client{restClient: client}
}

type Client struct {
	restClient *rest.Client
}

func (c Client) List(ctx context.Context) (api.Response, error) {
	path := basePath.String() + ":lean"
	resp, err := c.restClient.GET(ctx, path, rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           url.Values{"add-fields": []string{"EXTERNALID"}},
	})

	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	apiResp, err := api.NewResponseFromHTTPResponse(resp)
	if err != nil {
		return apiResp, err
	}

	if apiResp.Data, err = modifyBody(apiResp.Data); err != nil {
		return apiResp, api.RuntimeError{Resource: resource, Reason: "body transformation failed", Wrapped: err}
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
		return api.Response{}, api.ValidationError{Field: "id", Reason: "is empty"}
	}

	path := basePath.JoinPath(id).String()
	resp, err := c.restClient.GET(ctx, path, rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           url.Values{"add-fields": []string{"INCLUDES", "VARIABLES", "EXTERNALID", "RESOURCECONTEXT"}},
	})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func (c Client) Create(ctx context.Context, body []byte) (api.Response, error) {
	resp, err := c.restClient.POST(ctx, endpointPath, bytes.NewReader(body), rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPost, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func (c Client) Update(ctx context.Context, id string, body []byte) (api.Response, error) {
	if id == "" {
		return api.Response{}, api.ValidationError{Field: "id", Reason: "is empty"}
	}

	existing, err := c.Get(ctx, id)
	if err != nil {
		return api.Response{}, err
	}

	var getResponse struct {
		Version int    `json:"version"`
		Owner   string `json:"owner"`
	}
	err = json.Unmarshal(existing.Data, &getResponse)
	if err != nil || getResponse.Version == 0 || getResponse.Owner == "" {
		return api.Response{}, api.ValidationError{Field: "version, owner", Reason: "at least one is invalid"}
	}

	// Adds owner if not set(they are mandatory from the API),
	// and always override uid with the uid that is getting updated
	body, err = addOwnerAndUIDIfNotSet(body, getResponse.Owner, id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to add owner and UID", Wrapped: err}
	}

	path := basePath.JoinPath(id).String()
	resp, err := c.restClient.PUT(ctx, path, bytes.NewReader(body), rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           map[string][]string{"optimistic-locking-version": {strconv.Itoa(getResponse.Version)}}})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPut, Identifier: id, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func (c Client) Delete(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, api.ValidationError{Field: "id", Reason: "is empty"}
	}

	path := basePath.JoinPath(id).String()
	resp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodDelete, Identifier: id, Wrapped: err}
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
		return nil, api.RuntimeError{Resource: resource, Reason: "unmarshalling failed", Wrapped: err}
	}

	var result []api.Response
	for _, s := range segments {
		resp, err := c.Get(ctx, s.Uid)
		if err != nil {
			return nil, api.ClientError{Resource: resource, Identifier: s.Uid, Operation: http.MethodGet, Wrapped: err}
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
