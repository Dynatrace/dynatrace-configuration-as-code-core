/*
 * @license
 * Copyright 2023 Dynatrace LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

// Resource specifies information about a specific resource
type Resource struct {
	// Path is the API path to be used for this resource
	Path string
}

// ResourceType enumerates different kind of resources
type ResourceType int

const (
	Workflows ResourceType = iota
	BusinessCalendars
	SchedulingRules
)

var resources = map[ResourceType]Resource{
	Workflows:         {Path: "/platform/automation/v1/workflows"},
	BusinessCalendars: {Path: "/platform/automation/v1/business-calendars"},
	SchedulingRules:   {Path: "/platform/automation/v1/scheduling-rules"},
}

type listResponse struct {
	api.Response
	Count   int               `json:"count"`
	Objects []json.RawMessage `json:"results"`
}

// NewClient creates and returns a new instance a client which is used for interacting
// with automation resources.
//
// Parameters:
//
//   - client: A REST client used for making HTTP requests to interact with automation resources.
//
// Returns:
//
//   - *Client: A new instance of the Client type initialized with the provided REST client and resources.
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

// Get retrieves a single automation object based on the specified resource type and ID.
//
// It checks if the ID is non-empty, and if not, returns an error.
//
// Parameters:
//
//   - ctx: The context for the HTTP request.
//   - resourceType: The type of the resource to retrieve.
//   - id: The unique identifier of the object to retrieve.
//
// Returns:
//
//   - Response: A Response containing the result of the HTTP operation, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (a Client) Get(ctx context.Context, resourceType ResourceType, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf("id must be non empty")
	}
	path, err := url.JoinPath(resources[resourceType].Path, id)
	if err != nil {
		return api.Response{}, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := a.makeRequestWithAdminAccess(resourceType, func(options rest.RequestOptions) (*http.Response, error) {
		return a.restClient.GET(ctx, path, options)
	})

	if err != nil {
		return api.Response{}, fmt.Errorf("failed to get automation resource of type %q with id %q: %w", resourceType, id, err)
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// Create creates a new automation object based on the specified resource type.
//
// Parameters:
//
//   - ctx: The context for the HTTP request.
//   - resourceType: The type of the resource to retrieve.
//   - data: the data of the resource
//
// Returns:
//
//   - Response: A Response containing the result of the HTTP operation, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (a Client) Create(ctx context.Context, resourceType ResourceType, data []byte) (api.Response, error) {
	resp, err := a.makeRequestWithAdminAccess(resourceType, func(options rest.RequestOptions) (*http.Response, error) {
		return a.restClient.POST(ctx, resources[resourceType].Path, bytes.NewReader(data), options)
	})
	if err != nil {
		return api.Response{}, err
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// Update updates an automation object based on the specified resource type and id
//
// Parameters:
//
//   - ctx: The context for the HTTP request.
//   - resourceType: The type of the resource to retrieve.
//   - id: the id of the resource to be updated
//   - data: the updated data
//
// Returns:
//
//   - Response: A Response containing the result of the HTTP operation, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (a Client) Update(ctx context.Context, resourceType ResourceType, id string, data []byte) (api.Response, error) {
	if err := rmIDField(&data); err != nil {
		return api.Response{}, fmt.Errorf("unable to remove id field from payload in order to update object with ID %s: %w", id, err)
	}
	if id == "" {
		return api.Response{}, errors.New("id must be non empty")
	}
	path, err := url.JoinPath(resources[resourceType].Path, id)
	if err != nil {
		return api.Response{}, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := a.makeRequestWithAdminAccess(resourceType, func(options rest.RequestOptions) (*http.Response, error) {
		return a.restClient.PUT(ctx, path, bytes.NewReader(data), options)
	})
	if err != nil {
		return api.Response{}, err
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// List retrieves a list of automation objects of the specified resource
//
// The function sends multiple HTTP GET requests to fetch paginated data. It continues making requests
// until the total number of objects retrieved matches the expected count provided by the server.
//
// The result is returned as a slice of ListResponse objects. Each ListResponse contains information about
// the HTTP response, including the status code, response data, and request details. The objects retrieved
// from each paginated request are stored as byte slices within the ListResponse.
//
// Parameters:
//   - ctx: A context.Context for controlling the request lifecycle.
//   - resourceType: The type of resource to list.
//
// Returns:
//
//   - ListResponse: A ListResponse which is an api.PagedListResponse containing all objects fetched from the api
//   - error: An error if the HTTP call fails or another error happened.
func (a Client) List(ctx context.Context, resourceType ResourceType) (api.PagedListResponse, error) {
	// retVal are the collected paginated API responses
	var retVal api.PagedListResponse

	// result is the latest API response received
	var result listResponse
	result.Count = 1 // ensure starting condition is met, after the first API call this will be actual total count of objects
	retrieved := 0

	wfAdminAccess := resourceType == Workflows // only use admin access for workflows

	for retrieved < result.Count {
		opts := rest.RequestOptions{
			QueryParams: url.Values{
				"offset": []string{strconv.Itoa(retrieved)},
			},
		}
		if wfAdminAccess {
			opts.QueryParams["adminAccess"] = []string{"true"}
		}

		resp, err := a.restClient.GET(ctx, resources[resourceType].Path, opts)
		if err != nil {
			return api.PagedListResponse{}, fmt.Errorf("failed to list automation resources: %w", err)
		}

		// if Workflow API rejected the initial request with admin permissions -> retry without
		if wfAdminAccess && resp.StatusCode == http.StatusForbidden {
			wfAdminAccess = false
			resp.Body.Close()
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return api.PagedListResponse{}, api.NewAPIErrorFromResponseAndBody(resp, body)
		}

		// if one of the response has code != 2xx return empty list with response info
		if !rest.IsSuccess(resp) {
			return api.PagedListResponse{}, api.NewAPIErrorFromResponseAndBody(resp, body)
		}

		result, err = unmarshalJSONList(resp)
		if err != nil {
			return api.PagedListResponse{}, api.NewAPIErrorFromResponseAndBody(resp, body)
		}

		retrieved += len(result.Objects)

		b := make([][]byte, len(result.Objects))
		for i, v := range result.Objects {
			b[i], _ = v.MarshalJSON() // marshalling the JSON back to JSON will not fail
		}

		retVal = append(retVal, api.ListResponse{
			Response: api.Response{
				StatusCode: result.StatusCode,
				Data:       result.Data,
				Request:    result.Request,
			},
			Objects: b,
		})

	}

	return retVal, nil
}

// Upsert creates or updates an object of a specified resource type with the given ID and data.
//
// Parameters:
//   - ctx: The context for the HTTP request.
//   - resourceType: The type of the resource to upsert.
//   - id: The unique identifier for the object.
//   - data: The data payload representing the object.
//
// Returns:
//
//   - Response: A Response containing the result of the HTTP operation, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (a Client) Upsert(ctx context.Context, resourceType ResourceType, id string, data []byte) (api.Response, error) {
	resp, err := a.Update(ctx, resourceType, id, data)
	if err != nil {
		if api.IsNotFoundError(err) {
			return a.createWithID(ctx, resourceType, id, data)
		}
		return api.Response{}, err
	}
	return resp, nil
}

func (a Client) createWithID(ctx context.Context, resourceType ResourceType, id string, data []byte) (api.Response, error) {
	// make sure actual "id" field is set in payload
	if err := setIDField(id, &data); err != nil {
		return api.Response{}, fmt.Errorf("unable to set the id field in order to crate object with id %s: %w", id, err)
	}

	return a.Create(ctx, resourceType, data)
}

func (a Client) makeRequestWithAdminAccess(resourceType ResourceType, request func(options rest.RequestOptions) (*http.Response, error)) (*http.Response, error) {
	if resourceType == Workflows {
		opts := rest.RequestOptions{
			QueryParams: url.Values{"adminAccess": []string{"true"}},
			CustomShouldRetryFunc: func(resp *http.Response) bool {
				return rest.RetryIfNotSuccess(resp) && (resp.StatusCode != http.StatusForbidden)
			},
		}
		resp, err := request(opts)
		if err != nil {
			return nil, err
		}
		if resp != nil && resp.StatusCode == http.StatusForbidden {
			return request(rest.RequestOptions{})
		}
		return resp, err
	}

	return request(rest.RequestOptions{})
}

// Delete removes an automation object of the specified resource type by its unique identifier (ID).
//
// If the initial DELETE request results in a forbidden status code (HTTP 403) for Workflows, it retries
// the request without the "adminAccess" parameter.
//
// Parameters:
//   - ctx: A context.Context for controlling the request lifecycle.
//   - resourceType: The type of resource from which to delete the object.
//   - id: The unique identifier (ID) of the object to delete.
//
// Returns:
//
//   - Response: A Response containing the result of the HTTP operation, including status code and data.
//   - error: An error if the HTTP call fails or another error happened.
func (a Client) Delete(ctx context.Context, resourceType ResourceType, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, errors.New("id must be non empty")
	}
	path, err := url.JoinPath(resources[resourceType].Path, id)
	if err != nil {
		return api.Response{}, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := a.makeRequestWithAdminAccess(resourceType, func(options rest.RequestOptions) (*http.Response, error) {
		opts := rest.RequestOptions{
			QueryParams: options.QueryParams,
			CustomShouldRetryFunc: func(resp *http.Response) bool {
				if options.CustomShouldRetryFunc == nil {
					return rest.RetryOnFailureExcept404(resp)
				}

				return options.CustomShouldRetryFunc(resp) && rest.RetryOnFailureExcept404(resp)
			},
		}
		return a.restClient.DELETE(ctx, path, opts)
	})

	if err != nil {
		return api.Response{}, err
	}

	return api.NewResponseFromHTTPResponse(resp)
}

func unmarshalJSONList(raw *http.Response) (listResponse, error) {
	var r listResponse

	body, err := io.ReadAll(raw.Body)
	if err != nil {
		return r, err
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return listResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	r.Data = body
	r.StatusCode = raw.StatusCode
	r.Request = rest.RequestInfo{
		Method: raw.Request.Method,
		URL:    raw.Request.URL.String(),
	}
	return r, nil
}

func setIDField(id string, data *[]byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(*data, &m)
	if err != nil {
		return err
	}
	m["id"] = id
	*data, err = json.Marshal(m)
	if err != nil {
		return err
	}
	return nil
}

func rmIDField(data *[]byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(*data, &m)
	if err != nil {
		return err
	}
	delete(m, "id")
	*data, err = json.Marshal(m)
	if err != nil {
		return err
	}
	return nil
}
