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

package documents

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/documents"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

type DocumentType string

const (
	Dashboard DocumentType = "dashboard"
	Notebook  DocumentType = "notebook"
)

// Client is the HTTP client to be used for interacting with the Document API
type Client struct {
	client client
}

// NewClient creates a new document client
func NewClient(client *rest.Client) *Client {
	c := &Client{client: documents.NewClient(client)}
	return c
}

// Response contains the API response
type Response struct {
	api.Response
	Metadata
}

// ListResponse is a list of API Responses
type ListResponse struct {
	api.Response
	Responses []Response
}

func (c Client) Get(ctx context.Context, id string) (*Response, error) {
	resp, err := api.AsResponseOrError(c.client.Get(ctx, id))

	boundary, err := extractBoundary(resp)
	if err != nil {
		return nil, err
	}

	reader := multipart.NewReader(bytes.NewReader(resp.Data), boundary)

	form, err := reader.ReadForm(0)
	if err != nil {
		return nil, fmt.Errorf("unable to read multipart form: %w", err)
	}

	if len(form.Value["metadata"]) == 0 {
		return nil, fmt.Errorf("metadata field not found in response")
	}

	m := Metadata{}
	err = json.Unmarshal([]byte(form.Value["metadata"][0]), &m)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal metadata: %w", err)
	}

	file, err := form.File["content"][0].Open()
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	resp.Data, err = io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	return &Response{Response: *resp, Metadata: m}, nil
}

func extractBoundary(resp *api.Response) (string, error) {
	t, ps, err := mime.ParseMediaType(resp.Header.Get("content-type"))
	if !strings.HasPrefix(t, "multipart") {
		return "", errors.New("http response is not multipart")
	}
	if err != nil {
		return "", err
	}
	return ps["boundary"], nil
}

func (c Client) List(ctx context.Context, filter string) (*ListResponse, error) {
	type listResponse struct {
		TotalCount  int        `json:"totalCount"`
		Documents   []Response `json:"documents"`
		NextPageKey *string    `json:"nextPageKey"`
	}

	var retVal ListResponse
	var result listResponse
	var initialPage = ""
	result.NextPageKey = &initialPage

	for result.NextPageKey != nil {

		queryParams := url.Values{"filter": {filter}}
		if *result.NextPageKey != "" {
			queryParams.Add("page-key", *result.NextPageKey)
		}

		ro := rest.RequestOptions{
			QueryParams: queryParams,
		}

		resp, err := api.AsResponseOrError(c.client.List(ctx, ro))
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(resp.Data, &result)
		if err != nil {
			return nil, err
		}

		for i := range result.Documents {
			result.Documents[i].Request = rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL}
			result.Documents[i].StatusCode = resp.StatusCode
		}

		retVal.Responses = append(retVal.Responses, result.Documents...)
		retVal.StatusCode = resp.StatusCode
	}

	return &retVal, nil
}

func (c Client) Create(ctx context.Context, name string, isPrivate bool, externalId string, data []byte, documentType DocumentType) (*api.Response, error) {
	d := documents.Document{
		Kind:       string(documentType),
		Name:       name,
		Public:     !isPrivate,
		ExternalID: externalId,
		Content:    data,
	}

	resp, err := c.create(ctx, d)
	if err != nil {
		return nil, err
	}

	var md Metadata
	if md, err = UnmarshallMetadata(resp.Data); err != nil {
		return nil, err
	}

	r, err := c.patch(ctx, md.ID, md.Version, d)
	if err != nil {
		if !isNotFoundError(err) {
			if _, err1 := c.delete(ctx, md.ID, md.Version); err1 != nil {
				return nil, errors.Join(err, err1)
			}
		}
		return nil, err
	}
	return r, nil
}

func isNotFoundError(err error) bool {
	var apiErr api.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

func (c Client) Update(ctx context.Context, id string, name string, isPrivate bool, data []byte, documentType DocumentType) (*api.Response, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}

	resp, err := c.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	d := documents.Document{
		Kind:    string(documentType),
		Name:    name,
		Public:  !isPrivate,
		Content: data,
	}

	return c.patch(ctx, id, resp.Version, d)
}

func (c Client) Delete(ctx context.Context, id string) (*api.Response, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}

	resp, err := c.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return c.delete(ctx, id, resp.Version)
}

func (c Client) create(ctx context.Context, d documents.Document) (*api.Response, error) {
	return api.AsResponseOrError(c.client.Create(ctx, d))
}

func (c Client) patch(ctx context.Context, id string, version int, d documents.Document) (*api.Response, error) {
	resp, err := api.AsResponseOrError(c.client.Patch(ctx, id, version, d))
	if err != nil {
		return resp, err
	}

	tmp, err := extractMetadata(resp.Data)
	if err != nil {
		return resp, fmt.Errorf("extracting metadata failed: %w", err)
	}
	resp.Data = tmp

	return resp, nil
}

func (c Client) delete(ctx context.Context, id string, version int) (*api.Response, error) {
	r, err := api.AsResponseOrError(c.client.Delete(ctx, id, version))
	if err != nil {
		return r, err
	}

	return api.AsResponseOrError(c.client.Trash(ctx, id))
}

func extractMetadata(in []byte) (out []byte, err error) {
	var metadata map[string]any
	if err = json.Unmarshal(in, &metadata); err != nil {
		return
	}
	return json.Marshal(metadata["documentMetadata"])
}
