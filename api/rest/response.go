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

import "net/http"

// RequestInfo holds information about the original request that led to this response
type RequestInfo struct {
	Method string `json:"method"` // The HTTP method of the request.
	URL    string `json:"url"`    // The full URL of the request
}

func NewRequestInfoFromResponse(request *http.Request) RequestInfo {
	var method, url string
	if request != nil {
		method = request.Method
		if request.URL != nil {
			url = request.URL.String()
		}
	}
	return RequestInfo{Method: method, URL: url}
}

func IsSuccess(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode <= 299
}

func Is4xxError(resp *http.Response) bool {
	return resp.StatusCode >= 400 && resp.StatusCode <= 499
}

func Is5xxError(resp *http.Response) bool {
	return resp.StatusCode >= 500 && resp.StatusCode <= 599
}
