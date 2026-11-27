// @license
// Copyright 2026 Dynatrace LLC
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

// Package synthetic provides a client for the platform Synthetic v1 API.
package synthetic

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

// basePath is the root path of the platform Synthetic v1 API.
const basePath = "platform/synthetic/v1"

// NewClient creates a new Synthetic client using the given rest.Client.
func NewClient(client *rest.Client) *Client {
	return &Client{restClient: client}
}

// Client provides access to the platform Synthetic v1 API.
type Client struct {
	restClient *rest.Client
}

// processListResponse converts an HTTP response of a list endpoint into an api.ListResponse,
// extracting the objects found under arrayKey.
func processListResponse(httpResponse *http.Response, resource, arrayKey string) (api.ListResponse, error) {
	defer httpResponse.Body.Close()

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return api.ListResponse{}, api.NewAPIErrorFromResponse(httpResponse)
	}

	if !rest.IsSuccess(httpResponse) {
		return api.ListResponse{}, api.NewAPIErrorFromResponseAndBody(httpResponse, body)
	}

	// Decode only the array under arrayKey, so other top-level fields don't break the list.
	var s map[string]json.RawMessage
	if err := json.Unmarshal(body, &s); err != nil {
		return api.ListResponse{}, api.RuntimeError{Resource: resource, Reason: "failed to unmarshal list response", Wrapped: err}
	}

	var items []json.RawMessage
	if raw, ok := s[arrayKey]; ok {
		if err := json.Unmarshal(raw, &items); err != nil {
			return api.ListResponse{}, api.RuntimeError{Resource: resource, Reason: "failed to unmarshal list response", Wrapped: err}
		}
	}

	var data [][]byte
	for _, it := range items {
		data = append(data, it)
	}

	return api.ListResponse{
		StatusCode: httpResponse.StatusCode,
		Header:     httpResponse.Header,
		Data:       nil,
		Request:    api.NewRequestInfoFromRequest(httpResponse.Request),
		Objects:    data,
	}, nil
}
