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

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

// Response represents an API response
type Response struct {
	StatusCode int `json:"-"`
	Header     http.Header
	Data       []byte           `json:"-"`
	Request    rest.RequestInfo `json:"-"`
}

// AsResponseOrError is a helper function to convert an http.Response or error to a Response or error.
// It ensures that the response body is always read to completion and closed.
// Any non-successful (i.e. not 2xx) status code results in an APIError.
func AsResponseOrError(resp *http.Response, err error) (*Response, error) {
	var responseBody []byte
	var readErr error
	if resp != nil {
		responseBody, readErr = io.ReadAll(resp.Body)
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	if readErr != nil {
		return nil, NewAPIErrorFromResponseAndBody(resp, responseBody)
	}

	if !rest.IsSuccess(resp) {
		return nil, NewAPIErrorFromResponseAndBody(resp, responseBody)
	}

	response := NewResponseFromHTTPResponseAndBody(resp, responseBody)
	return &response, nil
}

func NewResponseFromHTTPResponseAndBody(resp *http.Response, body []byte) Response {
	return Response{
		Header:     resp.Header,
		StatusCode: resp.StatusCode,
		Data:       body,
		Request:    NewRequestInfoFromRequest(resp.Request),
	}
}

func NewRequestInfoFromRequest(request *http.Request) rest.RequestInfo {
	var method, url string
	if request != nil {
		method = request.Method
		if request.URL != nil {
			url = request.URL.String()
		}
	}
	return rest.RequestInfo{Method: method, URL: url}
}

// PagedListResponse is a list of ListResponse values.
// It is used by return values of APIs that support pagination.
// Each ListResponse entry possibly contains multiple objects of the fetched resource type.
// To get all response objects in a single slice of []byte you can call All().
//
// In case of any individual API request being unsuccessful, PagedListResponse will contain only that failed ListResponse.
type PagedListResponse []ListResponse

// All returns all objects of a PagedListResponse in one slice
func (p PagedListResponse) All() [][]byte {
	var ret [][]byte
	for _, l := range []ListResponse(p) {
		for _, o := range l.Objects {
			ret = append(ret, o)
		}
	}
	return ret
}

// ListResponse represents a multi-object API response
// It contains both the full JSON Data, and a slice of Objects for more convenient access
type ListResponse struct {
	Response
	Objects [][]byte `json:"-"`
}

// IsSuccess returns true if the response indicates a successful HTTP status code.
// A status code between 200 and 299 (inclusive) is considered a success.
func (r Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode <= 299
}

// Is4xxError returns true if the response indicates a 4xx client error HTTP status code.
// A status code between 400 and 499 (inclusive) is considered a client error.
func (r Response) Is4xxError() bool {
	return r.StatusCode >= 400 && r.StatusCode <= 499
}

// Is5xxError returns true if the response indicates a 5xx server error HTTP status code.
// A status code between 500 and 599 (inclusive) is considered a server error.
func (r Response) Is5xxError() bool {
	return r.StatusCode >= 500 && r.StatusCode <= 599
}

// AsAPIError converts a Response object to an APIError if it represents a 4xx or 5xx error.
// If the Response does not represent an error, it returns an empty APIError and false.
//
// Parameters:
// - r (Response): The Response object to convert to an APIError.
//
// Returns:
//   - (APIError, bool): An APIError containing error information and a boolean indicating
//     whether the conversion was successful (true for errors, false otherwise).
func (r Response) AsAPIError() (APIError, bool) {
	if r.Is4xxError() || r.Is5xxError() {
		return APIError{
			Body:       r.Data,
			StatusCode: r.StatusCode,
			Request:    r.Request,
		}, true
	}
	return APIError{}, false
}

// AsAPIError converts a PagedListResponse pointer to an APIError if it represents a 4xx or 5xx error.
// If the PagedListResponse does not represent an error, it returns an empty APIError and false.
// Unlike a normal Response, a PagedListResponse represents several individual API responses as a slice.
// In case of any response being unsuccessful, PagedListResponse will contain only that response by convention - this fact
// is used by this method.
//
// Parameters:
// - [ (PagedListResponse): The PagedListResponse object to convert to an APIError.
//
// Returns:
//   - (APIError, bool): An APIError containing error information and a boolean indicating
//     whether the conversion was successful (true for errors, false otherwise).
func (p PagedListResponse) AsAPIError() (APIError, bool) {
	if len(p) != 1 {
		return APIError{}, false
	}

	r := []ListResponse(p)[0]

	if r.Is4xxError() || r.Is5xxError() {
		return APIError{
			Body:       r.Data,
			StatusCode: r.StatusCode,
			Request:    r.Request,
		}, true
	}
	return APIError{}, false
}

// APIError represents an error returned by an API with associated information.
type APIError struct {
	StatusCode int              `json:"statusCode"` // StatusCode is the HTTP response status code returned by the API.
	Body       []byte           `json:"body"`       // Body is the HTTP payload returned by the API.
	Request    rest.RequestInfo `json:"request"`    // Request is information about the original request that led to this response error.
}

func NewAPIErrorFromResponseAndBody(resp *http.Response, body []byte) APIError {
	return APIError{
		StatusCode: resp.StatusCode,
		Body:       body,
		Request:    NewRequestInfoFromRequest(resp.Request),
	}
}

func NewAPIErrorFromResponse(resp *http.Response) error {
	apiErr := APIError{
		StatusCode: resp.StatusCode,
		Request:    rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()},
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Join(apiErr, fmt.Errorf("unable to read API response body: %w", err))
	}
	apiErr.Body = body
	return apiErr
}

// Error returns a string representation of the APIError, providing details about the failed API request.
// It includes the HTTP method, URL, status code, and response body.
//
// Returns:
// - string: A string representing the error message.
func (r APIError) Error() string {
	return fmt.Sprintf("API request HTTP %s %s failed with status code %d: %s", r.Request.Method, r.Request.URL, r.StatusCode, string(r.Body))
}

func (r APIError) Is4xxError() bool {
	return r.StatusCode >= 400 && r.StatusCode <= 499
}

func (r APIError) Is5xxError() bool {
	return r.StatusCode >= 500 && r.StatusCode <= 599
}

// DecodeJSON tries to unmarshal the Response.Data of the given Response r into an object of T.
func DecodeJSON[T any](r Response) (T, error) {
	var t T
	if err := json.Unmarshal(r.Data, &t); err != nil {
		return t, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return t, nil
}

// DecodeJSONObjects unmarshalls Objects contained in the given ListResponse into a slice of T.
// To decode the full JSON response contained in ListResponse use Response.DecodeJSON.
func DecodeJSONObjects[T any](r ListResponse) ([]T, error) {
	res := make([]T, len(r.Objects))
	for i, o := range r.Objects {
		var t T
		if err := json.Unmarshal(o, &t); err != nil {
			return []T{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		res[i] = t
	}

	return res, nil
}

// DecodePaginatedJSONObjects unmarshalls all objects contained in the given PagedListResponse into a slice of T.
// Alternative ways to access data are to use PagedListResponse as a []ListResponse and decode each ListResponse or
// to access and decode the entries as []byte via PagedListResponse.All.
func DecodePaginatedJSONObjects[T any](p PagedListResponse) ([]T, error) {
	var res []T
	for _, r := range []ListResponse(p) {
		ts, err := DecodeJSONObjects[T](r)
		if err != nil {
			return []T{}, err
		}
		res = append(res, ts...)
	}

	return res, nil
}
