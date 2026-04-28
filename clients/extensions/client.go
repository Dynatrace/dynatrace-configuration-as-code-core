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

package extensions

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	extensionsResourcePath           = "/platform/extensions/v2/extensions"
	monitoringResourcePath           = "monitoring-configurations"
	environmentConfigurationPath     = "environment-configuration"
	extensionsResource               = "extensions"
	monitoringConfigurationsResource = "monitoring-configurations"
	environmentConfigurationResource = "environment-configuration"
	urlCreationErrMsg                = "failed to construct URL"
	extensionsPageSize               = 100
	extensionVersionsPageSize        = 100
	monitoringConfigurationsPageSize = 500
)

var (
	extensionNameValidationErr   = api.ValidationError{Resource: extensionsResource, Field: "extension-name", Reason: "is empty"}
	configurationIDValidationErr = api.ValidationError{Resource: monitoringConfigurationsResource, Field: "configuration-id", Reason: "is empty"}
)

// Client is used to interact with the Extensions API.
type Client struct {
	restClient *rest.Client
}

// NewClient creates a new extensions Client using the given rest.Client.
func NewClient(client *rest.Client) *Client {
	return &Client{restClient: client}
}

// ListExtensions returns all extensions.
func (c Client) ListExtensions(ctx context.Context) (api.PagedListResponse, error) {
	var pagedListResponse api.PagedListResponse

	nextPageKey := ""
	for {
		var listResponse api.ListResponse
		var err error

		nextPageKey, listResponse, err = c.listExtensionsPage(ctx, nextPageKey)
		if err != nil {
			return nil, err
		}

		pagedListResponse = append(pagedListResponse, listResponse)
		if nextPageKey == "" {
			break
		}
	}

	return pagedListResponse, nil
}

func (c Client) listExtensionsPage(ctx context.Context, pageKey string) (string, api.ListResponse, error) {
	var ro rest.RequestOptions
	if pageKey != "" {
		ro.QueryParams = url.Values{"next-page-key": {pageKey}}
	} else {
		ro.QueryParams = url.Values{"page-size": {strconv.Itoa(extensionsPageSize)}}
	}

	resp, err := c.restClient.GET(ctx, extensionsResourcePath, ro)
	if err != nil {
		return "", api.ListResponse{}, api.ClientError{Resource: extensionsResource, Operation: http.MethodGet, Wrapped: err}
	}

	return processListResponse(resp, extensionsResource)
}

// ListExtensionVersions returns all installed versions of a given extension.
func (c Client) ListExtensionVersions(ctx context.Context, extensionName string) (api.PagedListResponse, error) {
	if extensionName == "" {
		return nil, extensionNameValidationErr
	}

	path, err := url.JoinPath(extensionsResourcePath, extensionName)
	if err != nil {
		return nil, api.RuntimeError{Resource: extensionsResource, Identifier: extensionName, Reason: urlCreationErrMsg, Wrapped: err}
	}
	return c.listAll(ctx, extensionName, path, extensionsResource, extensionVersionsPageSize)
}

// ListMonitoringConfigurations returns all monitoring configurations for a given extension.
func (c Client) ListMonitoringConfigurations(ctx context.Context, extensionName string) (api.PagedListResponse, error) {
	if extensionName == "" {
		return nil, extensionNameValidationErr
	}

	path, err := url.JoinPath(extensionsResourcePath, extensionName, monitoringResourcePath)
	if err != nil {
		return nil, api.RuntimeError{Resource: monitoringConfigurationsResource, Identifier: extensionName, Reason: urlCreationErrMsg, Wrapped: err}
	}

	return c.listAll(ctx, extensionName, path, monitoringConfigurationsResource, monitoringConfigurationsPageSize)
}

// listAll is a helper method to list paged resources.
// It takes care of paging through results until all pages have been retrieved and returns a combined PagedListResponse.
func (c Client) listAll(ctx context.Context, extensionName string, path string, resourceName string, pageSize int) (api.PagedListResponse, error) {
	var pagedListResponse api.PagedListResponse
	var nextPageKey string

	for {
		var listResponse api.ListResponse
		var err error

		var ro rest.RequestOptions
		if nextPageKey != "" {
			ro.QueryParams = url.Values{"next-page-key": {nextPageKey}}
		} else {
			ro.QueryParams = url.Values{"page-size": {strconv.Itoa(pageSize)}}
		}

		httpResp, err := c.restClient.GET(ctx, path, ro)
		if err != nil {
			return nil, api.ClientError{Resource: resourceName, Identifier: extensionName, Operation: http.MethodGet, Wrapped: err}
		}

		nextPageKey, listResponse, err = processListResponse(httpResp, resourceName)

		if err != nil {
			return nil, err
		}

		pagedListResponse = append(pagedListResponse, listResponse)
		if nextPageKey == "" {
			break
		}
	}
	return pagedListResponse, nil
}

// processListResponse is shared by both list endpoints. It unmarshals the "nextPageKey" and "items" fields.
func processListResponse(httpResponse *http.Response, resource string) (string, api.ListResponse, error) {
	resp, err := api.NewResponseFromHTTPResponse(httpResponse)
	if err != nil {
		return "", api.ListResponse{}, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	var listResp struct {
		NextPage string            `json:"nextPageKey"`
		Items    []json.RawMessage `json:"items"`
	}

	if err := json.Unmarshal(resp.Data, &listResp); err != nil {
		return "", api.ListResponse{}, api.RuntimeError{Resource: resource, Reason: "unmarshalling failed", Wrapped: err}
	}

	var objects [][]byte
	for _, it := range listResp.Items {
		objects = append(objects, it)
	}

	return listResp.NextPage,
		api.ListResponse{
			Response: api.Response{
				StatusCode: httpResponse.StatusCode,
				Header:     httpResponse.Header,
				Data:       nil,
				Request:    api.NewRequestInfoFromRequest(httpResponse.Request),
			},
			Objects: objects,
		},
		nil
}

// GetEnvironmentConfiguration returns the environment configuration for a given extension.
func (c Client) GetEnvironmentConfiguration(ctx context.Context, extensionName string) (api.Response, error) {
	if extensionName == "" {
		return api.Response{}, extensionNameValidationErr
	}

	path, err := url.JoinPath(extensionsResourcePath, extensionName, environmentConfigurationPath)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: environmentConfigurationResource, Identifier: extensionName, Reason: urlCreationErrMsg, Wrapped: err}
	}

	httpResp, err := c.restClient.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: environmentConfigurationResource, Identifier: extensionName, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: environmentConfigurationResource, Identifier: extensionName, Operation: http.MethodGet, Wrapped: err}
	}
	return resp, nil
}

// GetMonitoringConfiguration returns a specific monitoring configuration by extension name and configuration ID.
func (c Client) GetMonitoringConfiguration(ctx context.Context, extensionName string, configurationID string) (api.Response, error) {
	if extensionName == "" {
		return api.Response{}, extensionNameValidationErr
	}
	if configurationID == "" {
		return api.Response{}, configurationIDValidationErr
	}

	path, err := url.JoinPath(extensionsResourcePath, extensionName, monitoringResourcePath, configurationID)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: monitoringConfigurationsResource, Identifier: configurationID, Reason: urlCreationErrMsg, Wrapped: err}
	}

	httpResp, err := c.restClient.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: monitoringConfigurationsResource, Identifier: configurationID, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: monitoringConfigurationsResource, Identifier: configurationID, Operation: http.MethodGet, Wrapped: err}
	}
	return resp, nil
}

// CreateMonitoringConfiguration creates a new monitoring configuration for a given extension.
func (c Client) CreateMonitoringConfiguration(ctx context.Context, extensionName string, data []byte) (api.Response, error) {
	if extensionName == "" {
		return api.Response{}, extensionNameValidationErr
	}

	path, err := url.JoinPath(extensionsResourcePath, extensionName, monitoringResourcePath)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: monitoringConfigurationsResource, Identifier: extensionName, Reason: urlCreationErrMsg, Wrapped: err}
	}

	httpResp, err := c.restClient.POST(ctx, path, bytes.NewReader(data), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: monitoringConfigurationsResource, Identifier: extensionName, Operation: http.MethodPost, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: monitoringConfigurationsResource, Identifier: extensionName, Operation: http.MethodPost, Wrapped: err}
	}
	return resp, nil
}

// UpdateMonitoringConfiguration updates an existing monitoring configuration.
func (c Client) UpdateMonitoringConfiguration(ctx context.Context, extensionName string, configurationID string, data []byte) (api.Response, error) {
	if extensionName == "" {
		return api.Response{}, extensionNameValidationErr
	}
	if configurationID == "" {
		return api.Response{}, configurationIDValidationErr
	}

	path, err := url.JoinPath(extensionsResourcePath, extensionName, monitoringResourcePath, configurationID)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: monitoringConfigurationsResource, Identifier: configurationID, Reason: urlCreationErrMsg, Wrapped: err}
	}

	httpResp, err := c.restClient.PUT(ctx, path, bytes.NewReader(data), rest.RequestOptions{})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: monitoringConfigurationsResource, Identifier: configurationID, Operation: http.MethodPut, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: monitoringConfigurationsResource, Identifier: configurationID, Operation: http.MethodPut, Wrapped: err}
	}
	return resp, nil
}

// DeleteMonitoringConfiguration deletes a specific monitoring configuration.
func (c Client) DeleteMonitoringConfiguration(ctx context.Context, extensionName string, configurationID string) error {
	if extensionName == "" {
		return extensionNameValidationErr
	}
	if configurationID == "" {
		return configurationIDValidationErr
	}

	path, err := url.JoinPath(extensionsResourcePath, extensionName, monitoringResourcePath, configurationID)
	if err != nil {
		return api.RuntimeError{Resource: monitoringConfigurationsResource, Identifier: configurationID, Reason: urlCreationErrMsg, Wrapped: err}
	}

	httpResp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.ClientError{Resource: monitoringConfigurationsResource, Identifier: configurationID, Operation: http.MethodDelete, Wrapped: err}
	}

	_, err = api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.ClientError{Resource: monitoringConfigurationsResource, Identifier: configurationID, Operation: http.MethodDelete, Wrapped: err}
	}
	return nil
}
