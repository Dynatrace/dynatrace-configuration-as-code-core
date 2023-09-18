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
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"net/http"
	"net/url"
	"strconv"
)

type Response struct {
	api.Response
}

// PagedListResponse contains multiple ListResponse values
type PagedListResponse []ListResponse

// Objects returns all objects of a PagedListResponse
func (p *PagedListResponse) Objects() [][]byte {
	var ret [][]byte
	for _, l := range []ListResponse(*p) {
		for _, o := range l.Objects {
			ret = append(ret, o)
		}
	}
	return ret
}

type ListResponse struct {
	api.ListResponse
}

type listResponse struct {
	api.Response
	Count   int               `json:"count"`
	Objects []json.RawMessage `json:"results"`
}

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

// NewClient creates and returns a new instance a client which is used for interacting
// with automation resources.
//
// Parameters:
//
//   - client (*rest.Client): A REST client used for making HTTP requests to interact with automation resources.
//
// Returns:
//
//   - c (*Client): A new instance of the Client type initialized with the provided REST client and resources.
func NewClient(client *rest.Client) *Client {
	c := &Client{
		client:    client,
		resources: resources,
	}

	return c
}

// Client can be used to interact with the Automation API
type Client struct {
	client    *rest.Client
	resources map[ResourceType]Resource
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
//	ctx (context.Context): The context for the HTTP request.
//	resourceType (ResourceType): The type of the resource to retrieve.
//	id (string): The unique identifier of the object to retrieve.
//
// Returns:
//
//	result (Response): A Response object containing the retrieved object and its metadata.
//	err (error): An error if the retrieval operation fails or if the ID is empty.
func (a Client) Get(ctx context.Context, resourceType ResourceType, id string) (result Response, err error) {
	if id == "" {
		return Response{}, fmt.Errorf("id must be non empty")
	}
	path, err := url.JoinPath(a.resources[resourceType].Path, id)
	if err != nil {
		return Response{}, fmt.Errorf("failed to create URL: %w", err)
	}
	resp, err := a.client.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return Response{}, fmt.Errorf("failed to get automation resource of type %q with id %q: %w", resourceType, id, err)
	}

	return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload, Request: resp.RequestInfo}}, nil
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
//   - ctx (context.Context): A context.Context for controlling the request lifecycle.
//   - resourceType (ResourceType): The type of resource to list.
//
// Returns:
//
//	(Response, error): A Response object containing information about the HTTP response
//	and an error, if any.
func (a Client) List(ctx context.Context, resourceType ResourceType) (PagedListResponse, error) {

	var retVal []ListResponse
	var result listResponse
	result.Count = 1

	getCount := func(l []ListResponse) int {
		var count int
		for _, r := range l {
			count += len(r.Objects)
		}
		return count
	}

	for getCount(retVal) < result.Count {
		resp, err := a.client.GET(ctx, a.resources[resourceType].Path, rest.RequestOptions{
			QueryParams: url.Values{"offset": []string{strconv.Itoa(len(retVal))}}},
		)

		if err != nil {
			return []ListResponse{}, fmt.Errorf("failed to list buckets:%w", err)
		}

		// if one of the response has code != 2xx return empty list with response info
		if !resp.IsSuccess() {
			return []ListResponse{{
				api.ListResponse{
					Response: api.Response{
						StatusCode: resp.StatusCode,
						Data:       resp.Payload,
						Request:    resp.RequestInfo,
					},
					Objects: nil,
				},
			}}, nil
		}

		result, err = unmarshalJSONList(&resp)
		if err != nil {
			return []ListResponse{}, fmt.Errorf("failed to parse list response:%w", err)
		}

		b := make([][]byte, len(result.Objects))
		for i, v := range result.Objects {
			b[i], _ = v.MarshalJSON() // marshalling the JSON back to JSON will not fail
		}

		retVal = append(retVal, ListResponse{api.ListResponse{
			Response: api.Response{
				StatusCode: result.StatusCode,
				Data:       result.Data,
				Request:    result.Request,
			},
			Objects: b,
		}})

	}

	return retVal, nil

}

// Upsert creates or updates an object of a specified resource type with the given ID and data.
//
// Parameters:
//   - ctx (context.Context): The context for the HTTP request.
//   - resourceType (ResourceType): The type of the resource to upsert.
//   - id (string): The unique identifier for the object.
//   - data ([]byte): The data payload representing the object.
//
// Returns:
//
//	(Response, error): A Response object containing information about the HTTP response
//	and an error, if any.
func (a Client) Upsert(ctx context.Context, resourceType ResourceType, id string, data []byte) (result Response, err error) {
	if id == "" {
		return Response{}, fmt.Errorf("id must be non empty")
	}
	if err := rmIDField(&data); err != nil {
		return Response{}, fmt.Errorf("unable to remove id field from payload in order to update object with ID %s: %w", id, err)
	}

	path, err := url.JoinPath(a.resources[resourceType].Path, id)
	if err != nil {
		return Response{}, fmt.Errorf("failed to create URL: %w", err)
	}

	workflowsAdminAccess := resourceType == Workflows

	resp, err := a.client.PUT(ctx, path, bytes.NewReader(data), rest.RequestOptions{
		QueryParams: url.Values{"adminAccess": []string{strconv.FormatBool(workflowsAdminAccess)}},
	})
	if err != nil {
		return Response{}, fmt.Errorf("unable to update object with ID %s: %w", id, err)
	}

	if workflowsAdminAccess && resp.StatusCode == http.StatusForbidden {

		resp, err = a.client.PUT(ctx, path, bytes.NewReader(data), rest.RequestOptions{})
		if err != nil {
			return Response{}, fmt.Errorf("unable to update object with ID %s: %w", id, err)
		}
	}

	if resp.IsSuccess() {
		return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload, Request: resp.RequestInfo}}, nil
	}

	// at this point we need to create a new object using HTTP POST
	return a.create(ctx, id, data, resourceType)
}

func (a Client) create(ctx context.Context, id string, data []byte, resourceType ResourceType) (Response, error) {
	// make sure actual "id" field is set in payload
	if err := setIDField(id, &data); err != nil {
		return Response{}, fmt.Errorf("unable to set the id field in order to crate object with id %s: %w", id, err)
	}

	resp, err := a.client.POST(ctx, a.resources[resourceType].Path, bytes.NewReader(data), rest.RequestOptions{})
	if err != nil {
		return Response{}, err
	}

	return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload, Request: resp.RequestInfo}}, nil
}

// Delete removes an automation object of the specified resource type by its unique identifier (ID).
//
// If the initial DELETE request results in a forbidden status code (HTTP 403) for Workflows, it retries
// the request without the "adminAccess" parameter.
//
// Parameters:
//   - ctx (context.Context): A context.Context for controlling the request lifecycle.
//   - resourceType (ResourceType): The type of resource from which to delete the object.
//   - id (string): The unique identifier (ID) of the object to delete.
//
// Returns:
//
//	(Response, error): A Response object containing information about the HTTP response and an error, if any.
func (a Client) Delete(ctx context.Context, resourceType ResourceType, id string) (Response, error) {
	if id == "" {
		return Response{}, fmt.Errorf("id must be non empty")
	}
	path, err := url.JoinPath(a.resources[resourceType].Path, id)
	if err != nil {
		return Response{}, fmt.Errorf("failed to create URL: %w", err)
	}

	workflowsAdminAccess := resourceType == Workflows

	resp, err := a.client.DELETE(ctx, path, rest.RequestOptions{
		QueryParams: url.Values{"adminAccess": []string{strconv.FormatBool(workflowsAdminAccess)}},
	})
	if err != nil {
		return Response{}, fmt.Errorf("unable to delete object with ID %s: %w", id, err)
	}

	if workflowsAdminAccess && resp.StatusCode == http.StatusForbidden {
		resp, err = a.client.DELETE(ctx, path, rest.RequestOptions{})
		if err != nil {
			return Response{}, fmt.Errorf("unable to delete object with ID %s: %w", id, err)
		}
	}
	return Response{api.Response{StatusCode: resp.StatusCode, Data: resp.Payload, Request: resp.RequestInfo}}, nil
}

// unmarshalJSONList unmarshals JSON data into a listResponse struct.
func unmarshalJSONList(raw *rest.Response) (listResponse, error) {
	var r listResponse
	err := json.Unmarshal(raw.Payload, &r)
	if err != nil {
		return listResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	r.Data = raw.Payload
	r.StatusCode = raw.StatusCode
	r.Request = raw.RequestInfo
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
