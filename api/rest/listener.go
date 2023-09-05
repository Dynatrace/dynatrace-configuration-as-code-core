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

package rest

import (
	"net/http"
	"time"
)

// RequestResponse represents a recorded HTTP request and response.
type RequestResponse struct {
	ID        string         // ID to identify and correlate requests to responses
	Timestamp time.Time      // Timestamp of the recorded request/response
	Request   *http.Request  // HTTP request
	Response  *http.Response // HTTP response
	Error     error          // Error, if any, during request/response
}

// IsRequest checks if the RequestResponse value represents an HTTP request.
//
// It returns a pointer to an http.Request and true if the RequestResponse contains a request.
// If the RequestResponse does not contain a request, it returns nil and false.
func (r *RequestResponse) IsRequest() (*http.Request, bool) {
	if r.Request != nil {
		return r.Request, true
	}
	return nil, false
}

// IsResponse checks if the RequestResponse value represents an HTTP response.
//
// It returns a pointer to an http.Response and true if the RequestResponse contains a response.
// If the RequestResponse does not contain a response, it returns nil and false.
func (r *RequestResponse) IsResponse() (*http.Response, bool) {
	if r.Response != nil {
		return r.Response, true
	}
	return nil, false
}

// HTTPListener is a struct that can be used to listen for HTTP requests and responses
// and invoke a user-defined callback function.
type HTTPListener struct {
	Callback func(response RequestResponse)
}

// onRequest is a method of HTTPListener that is called when an HTTP request is received.
// It invokes the user-defined callback function with the request information.
func (r *HTTPListener) onRequest(id string, req *http.Request) {
	if r != nil {
		r.Callback(RequestResponse{ID: id, Timestamp: time.Now(), Request: req})
	}
}

// onResponse is a method of HTTPListener that is called when an HTTP response is received.
// It invokes the user-defined callback function with the response information.
func (r *HTTPListener) onResponse(id string, resp *http.Response, err error) {
	if r != nil {
		reqResp := RequestResponse{ID: id, Timestamp: time.Now(), Response: resp, Error: err}
		r.Callback(reqResp)
	}
}
