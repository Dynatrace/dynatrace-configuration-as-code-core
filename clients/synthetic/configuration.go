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

package synthetic

import (
	"bytes"
	"context"
	"net/http"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	configurationPath     = basePath + "/configuration"
	configurationResource = "synthetic configuration"
)

// GetConfiguration returns the tenant-wide synthetic configuration.
func (c *Client) GetConfiguration(ctx context.Context) (api.Response, error) {
	resp, err := c.restClient.GET(ctx, configurationPath, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: configurationResource, Operation: http.MethodGet, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}

// UpdateConfiguration updates the tenant-wide synthetic configuration.
func (c *Client) UpdateConfiguration(ctx context.Context, body []byte) (api.Response, error) {
	resp, err := c.restClient.PUT(ctx, configurationPath, bytes.NewReader(body), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: configurationResource, Operation: http.MethodPut, Wrapped: err}
	}

	return api.NewResponseFromHTTPResponse(resp)
}
