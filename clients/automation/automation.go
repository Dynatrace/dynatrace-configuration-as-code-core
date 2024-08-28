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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/automation"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const bodyReadErrMsg = "unable to read API response body"

type Response = api.Response

type ListResponse = api.PagedListResponse

type listResponse struct {
	Response
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
		client:     automation.NewClient(client),
		restClient: client,
	}

	return c
}

// Client can be used to interact with the Automation API
type Client struct {
	client     *automation.Client
	restClient *rest.Client
}

// ClientOption are (optional) additional parameter passed to the creation of
// an automation client
type ClientOption func(*Client)

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
func (a Client) Get(ctx context.Context, resourceType automation.ResourceType, id string) (*Response, error) {
	return api.AsResponseOrError(a.client.Get(ctx, resourceType, id))
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
func (a Client) Create(ctx context.Context, resourceType automation.ResourceType, data []byte) (*Response, error) {
	return api.AsResponseOrError(a.client.Create(ctx, resourceType, data))
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
func (a Client) Update(ctx context.Context, resourceType automation.ResourceType, id string, data []byte) (*Response, error) {
	if err := rmIDField(&data); err != nil {
		return nil, fmt.Errorf("unable to remove id field from payload in order to update object with ID %s: %w", id, err)
	}
	return api.AsResponseOrError(a.client.Update(ctx, resourceType, id, data))
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
func (a Client) List(ctx context.Context, resourceType automation.ResourceType) (ListResponse, error) {
	// retVal are the collected paginated API responses
	var retVal ListResponse

	// result is the latest API response received
	var result listResponse
	result.Count = 1 // ensure starting condition is met, after the first API call this will be actual total count of objects
	retrieved := 0

	wfAdminAccess := resourceType == automation.Workflows // only use admin access for workflows

	for retrieved < result.Count {
		opts := rest.RequestOptions{
			QueryParams: url.Values{
				"offset": []string{strconv.Itoa(retrieved)},
			},
		}
		if wfAdminAccess {
			opts.QueryParams["adminAccess"] = []string{"true"}
		}

		resp, err := a.restClient.GET(ctx, automation.Resources[resourceType].Path, opts)
		if err != nil {
			return ListResponse{}, fmt.Errorf("failed to list automation resources: %w", err)
		}

		// if Workflow API rejected the initial request with admin permissions -> retry without
		if wfAdminAccess && resp.StatusCode == http.StatusForbidden {
			wfAdminAccess = false
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
			return ListResponse{}, api.NewAPIErrorFromResponseAndBody(resp, body)
		}

		// if one of the response has code != 2xx return empty list with response info
		if !rest.IsSuccess(resp) {
			return ListResponse{}, api.NewAPIErrorFromResponseAndBody(resp, body)
		}

		result, err = unmarshalJSONList(resp)
		if err != nil {
			logr.FromContextOrDiscard(ctx).Error(err, "failed to unmarshal json response")
			return ListResponse{}, api.NewAPIErrorFromResponseAndBody(resp, body)
		}

		retrieved += len(result.Objects)

		b := make([][]byte, len(result.Objects))
		for i, v := range result.Objects {
			b[i], _ = v.MarshalJSON() // marshalling the JSON back to JSON will not fail
		}

		retVal = append(retVal, api.ListResponse{
			Response: Response{
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
func (a Client) Upsert(ctx context.Context, resourceType automation.ResourceType, id string, data []byte) (result *Response, err error) {
	resp, err := a.Update(ctx, resourceType, id, data)
	if err != nil {
		var apiErr api.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return a.createWithID(ctx, resourceType, id, data)
		}
		return nil, err
	}
	return resp, nil
}

func (a Client) createWithID(ctx context.Context, resourceType automation.ResourceType, id string, data []byte) (*Response, error) {
	// make sure actual "id" field is set in payload
	if err := setIDField(id, &data); err != nil {
		return nil, fmt.Errorf("unable to set the id field in order to crate object with id %s: %w", id, err)
	}

	return api.AsResponseOrError(a.client.Create(ctx, resourceType, data))
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
func (a Client) Delete(ctx context.Context, resourceType automation.ResourceType, id string) (*Response, error) {
	return api.AsResponseOrError(a.client.Delete(ctx, resourceType, id))
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
