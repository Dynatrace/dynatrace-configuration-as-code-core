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

package grailfiltersegments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/grailfiltersegements"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

type Response = api.Response

func NewClient(client *rest.Client) *Client {
	c := &Client{
		client: grailfiltersegements.NewClient(client),
	}
	return c
}

// Client can be used to interact with the Automation API
type Client struct {
	client client
}

//go:generate mockgen -source filtersegments.go -package=grailfiltersegments -destination=client_mock.go
type client interface {
	List(ctx context.Context, ro rest.RequestOptions) (*http.Response, error)
	Get(ctx context.Context, id string, ro rest.RequestOptions) (*http.Response, error)
	Create(ctx context.Context, data []byte, ro rest.RequestOptions) (*http.Response, error)
	Update(ctx context.Context, id string, data []byte, ro rest.RequestOptions) (*http.Response, error)
	Delete(ctx context.Context, id string, ro rest.RequestOptions) (*http.Response, error)
}

var _ client = (*grailfiltersegements.Client)(nil)

// List gets a complete set of available configs. The Data filed in response is normalized to json list of entries.
func (c Client) List(ctx context.Context) (Response, error) {
	resp, err := c.client.List(ctx, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	defer closeBody(resp)
	if err != nil {
		return Response{}, fmt.Errorf("failed to list filtersegments resources: %w", err)
	}
	return processResponse(resp, normalizeListResponse)
}

// normalizeListResponse transform received json response to contain just a JSON list of elements
func normalizeListResponse(source []byte) ([]byte, error) {
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

// Get gets a complete configuration of filter segment with an ID
func (c Client) Get(ctx context.Context, id string) (Response, error) {
	if id == "" {
		return Response{}, errors.New("missing required id")
	}

	resp, err := c.client.Get(ctx, id, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	defer closeBody(resp)
	if err != nil {
		return Response{}, fmt.Errorf("failed to get filtersegment resource with id %s: %w", id, err)
	}
	return processResponse(resp, nil)
}

// GetAll gets a complete set of complete configuration for all available filter segments
func (c Client) GetAll(ctx context.Context) ([]Response, error) {
	const errMsg = "failed to get all filter segments: %w"
	listResp, err := c.List(ctx)
	if err != nil {
		return nil, err
	}

	var segments []struct {
		Uid string `json:"uid"`
	}
	if err = json.Unmarshal(listResp.Data, &segments); err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	var result []Response
	for _, s := range segments {
		resp, err := c.Get(ctx, s.Uid)
		if err != nil {
			return nil, fmt.Errorf(errMsg, err)
		}
		result = append(result, resp)
	}

	return result, nil
}

// Upsert creates a new entry of filter segment on server in case a configuration with an ID doesn't already exist on server. If exists, it updates it.
func (c Client) Upsert(ctx context.Context, id string, data []byte) (Response, error) {
	const errMsg = "failed to upsert filter segments resource with id %s: %w"
	existing, err := c.client.Get(ctx, id, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	closeBody(existing)
	if err != nil {
		return Response{}, fmt.Errorf(errMsg, id, err)
	}

	if existing.StatusCode == http.StatusNotFound {
		resp, err := c.client.Create(ctx, data, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
		closeBody(resp)
		if err != nil {
			return Response{}, fmt.Errorf(errMsg, id, err)
		}
		return processResponse(resp, nil)
	}

	existingResourceBody, err := io.ReadAll(existing.Body)
	if err != nil {
		return Response{}, api.NewAPIErrorFromResponseAndBody(existing, existingResourceBody)
	}

	var currentVersion struct {
		Version int `json:"version"`
	}
	err = json.Unmarshal(existingResourceBody, &currentVersion)
	if err != nil {
		return Response{}, fmt.Errorf(errMsg, id, err)
	}
	if currentVersion.Version == 0 {
		return Response{}, fmt.Errorf("missing version field in API response")
	}

	updateResourceResp, err := c.client.Update(ctx, id, data, rest.RequestOptions{
		CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
		QueryParams:           map[string][]string{"optimistic-locking-version": {fmt.Sprint(currentVersion.Version)}}})
	closeBody(updateResourceResp)
	if err != nil {
		return Response{}, fmt.Errorf("failed to update filter segments resource with id %s and version %d: %w", id, currentVersion.Version, err)
	}

	return processResponse(updateResourceResp, nil)
}

func processResponse(httpResponse *http.Response, transform func([]byte) ([]byte, error)) (Response, error) {
	var body []byte
	var err error

	if body, err = io.ReadAll(httpResponse.Body); err != nil {
		return Response{}, api.NewAPIErrorFromResponseAndBody(httpResponse, body)
	}

	if !rest.IsSuccess(httpResponse) {
		return Response{}, api.NewAPIErrorFromResponseAndBody(httpResponse, body)
	}

	if transform != nil {
		if body, err = transform(body); err != nil {
			return Response{}, api.NewAPIErrorFromResponseAndBody(httpResponse, body)
		}
	}

	return api.NewResponseFromHTTPResponseAndBody(httpResponse, body), nil
}

// Delete removes configuration for filter segment with given ID from a server.
func (c Client) Delete(ctx context.Context, id string) (Response, error) {
	if id == "" {
		return Response{}, errors.New("missing required id")
	}

	resp, err := c.client.Delete(ctx, id, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	closeBody(resp)
	if err != nil {
		return Response{}, fmt.Errorf("failed to delete filtersegment resource with id %s: %w", id, err)
	}

	if !rest.IsSuccess(resp) {
		return Response{}, api.NewAPIErrorFromResponse(resp)
	}
	return api.NewResponseFromHTTPResponseAndBody(resp, nil), nil
}

func closeBody(httpResponse *http.Response) {
	if httpResponse != nil && httpResponse.Body != nil {
		_ = httpResponse.Body.Close()
	}
}
