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
	"github.com/go-logr/logr"
)

const bodyReadErrMsg = "unable to read API response body"

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
	Get(ctx context.Context, id string, ro rest.RequestOptions) (*http.Response, error)
	List(ctx context.Context) (*http.Response, error)
	Create(ctx context.Context, data []byte) (*http.Response, error)
	Update(ctx context.Context, id string, data []byte, ro rest.RequestOptions) (*http.Response, error)
	Delete(ctx context.Context, id string) (*http.Response, error)
}

var _ client = (*grailfiltersegements.Client)(nil)

func (c Client) Get(ctx context.Context, id string) (Response, error) {
	if id == "" {
		return Response{}, errors.New("missing required id")
	}
	resp, err := c.client.Get(ctx, id, rest.RequestOptions{CustomShouldRetryFunc: rest.RetryIfTooManyRequests})
	if err != nil {
		return Response{}, fmt.Errorf("failed to get filtersegment resource with id %s: %w", id, err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

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

func (c Client) List(ctx context.Context) (Response, error) {
	resp, err := c.client.List(ctx)
	if err != nil {
		return Response{}, fmt.Errorf("failed to list filtersegments resources: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
		return Response{}, api.NewAPIErrorFromResponseAndBody(resp, body)
	}
	if !rest.IsSuccess(resp) {
		return Response{}, api.NewAPIErrorFromResponseAndBody(resp, body)
	}
	{
		var tmp map[string]any
		if err = json.Unmarshal(body, &tmp); err != nil {
			return Response{}, api.NewAPIErrorFromResponseAndBody(resp, body)
		}
		if body, err = json.Marshal(tmp["filterSegments"]); err != nil {
			return Response{}, api.NewAPIErrorFromResponseAndBody(resp, body)
		}
	}

	return api.NewResponseFromHTTPResponseAndBody(resp, body), nil
}

func (c Client) GetAll(ctx context.Context) ([]Response, error) {
	listResp, err := c.List(ctx)
	if err != nil {
		return nil, err
	}

	type filterSegment struct {
		Uid string `json:"uid"`
	}

	var fsegments []filterSegment
	if err = json.Unmarshal(listResp.Data, &fsegments); err != nil {
		return nil, err
	}

	var result []Response
	for _, f := range fsegments {
		resp, err := c.client.Get(ctx, f.Uid, getRequestOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to get filter segments resource with id %s: %w", f.Uid, err)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
			return nil, api.NewAPIErrorFromResponseAndBody(resp, body)
		}
		if !rest.IsSuccess(resp) {
			return nil, api.NewAPIErrorFromResponseAndBody(resp, body)
		}

		result = append(result, api.NewResponseFromHTTPResponseAndBody(resp, body))
	}

	return result, nil

}

func (c Client) Upsert(ctx context.Context, uid string, data []byte) (Response, error) {
	existingResourceResp, err := c.client.Get(ctx, uid, rest.RequestOptions{})
	if err != nil {
		return Response{}, fmt.Errorf("failed to get filter segments resource with id %s: %w", uid, err)
	}

	if existingResourceResp.StatusCode == http.StatusNotFound {
		newResourceResp, err := c.client.Create(ctx, data)
		if err != nil {
			return Response{}, fmt.Errorf("failed to create filter segments resource: %w", err)
		}

		newResourceBody, err := io.ReadAll(newResourceResp.Body)
		if err != nil {
			logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
			return Response{}, api.NewAPIErrorFromResponseAndBody(newResourceResp, newResourceBody)
		}

		return api.NewResponseFromHTTPResponseAndBody(newResourceResp, newResourceBody), nil
	}

	existingResourceBody, err := io.ReadAll(existingResourceResp.Body)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
		return Response{}, api.NewAPIErrorFromResponseAndBody(existingResourceResp, existingResourceBody)
	}

	type respWithVersion struct {
		Version int `json:"version"`
	}

	var currentVersion respWithVersion
	err = json.Unmarshal(existingResourceBody, &currentVersion)
	if err != nil {
		return Response{}, fmt.Errorf("unable to unmarshal data: %w", err)
	}
	if currentVersion.Version == 0 {
		return Response{}, fmt.Errorf("missing version field in API response")
	}

	updateResourceResp, err := c.client.Update(ctx, uid, data, rest.RequestOptions{QueryParams: map[string][]string{
		"optimistic-locking-version": {fmt.Sprint(currentVersion.Version)},
	}})

	if err != nil {
		return Response{}, fmt.Errorf("failed to update filter segments resource with id %s and version %d: %w", uid, currentVersion.Version, err)
	}

	updateResourceBody, err := io.ReadAll(updateResourceResp.Body)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
		return Response{}, api.NewAPIErrorFromResponseAndBody(updateResourceResp, updateResourceBody)
	}

	return api.NewResponseFromHTTPResponseAndBody(updateResourceResp, updateResourceBody), nil
}

func (c Client) Delete(ctx context.Context, id string) (Response, error) {
	if id == "" {
		return Response{}, errors.New("missing required id")
	}
	resp, err := c.client.Delete(ctx, id)
	if err != nil {
		return Response{}, fmt.Errorf("failed to get filtersegment resource with id %s: %w", id, err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if !rest.IsSuccess(resp) {
		return Response{}, api.NewAPIErrorFromResponse(resp)
	}
	return api.NewResponseFromHTTPResponse(resp), nil
}

var getRequestOptions = rest.RequestOptions{
	CustomShouldRetryFunc: rest.RetryIfTooManyRequests,
	QueryParams: map[string][]string{
		"add-fields": {"INCLUDES", "VARIABLES", "RESOURCECONTEXT"},
	},
}
