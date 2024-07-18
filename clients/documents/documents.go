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
	"github.com/go-logr/logr"
)

const bodyReadErrMsg = "unable to read API response body"

const optimisticLockingHeader = "optimistic-locking-version"

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

	// Metadata fields
	ID         string `json:"id"`
	ExternalID string `json:"externalId"`
	Actor      string `json:"actor"`
	Owner      string `json:"owner"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Version    int    `json:"version"`
	IsPrivate  bool   `json:"isPrivate"`
}

type Response2 = api.Response

type documentMetaData struct {
	ID         string `json:"id"`
	ExternalID string `json:"externalId"`
	Actor      string `json:"actor"`
	Owner      string `json:"owner"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Version    int    `json:"version"`
}

func metadata(b []byte) (documentMetaData, error) {
	// type metadata struct {
	// 	DocumentMetaData documentMetaData `json:"documentMetadata"`
	// }

	var m documentMetaData
	if err := json.Unmarshal(b, &m); err != nil {
		return documentMetaData{}, err
	}

	return m, nil
}

// ListResponse is a list of API Responses
type ListResponse struct {
	api.Response
	Responses []Response
}

func (c Client) Get(ctx context.Context, id string) (Response, error) {
	var r Response

	httpResp, err := c.client.Get(ctx, id)
	if err != nil {
		return Response{}, fmt.Errorf("failed to get document resource with id %s: %w", id, err)
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
		return Response{}, api.NewAPIErrorFromResponseAndBody(httpResp, body)
	}

	if err = httpResp.Body.Close(); err != nil {
		return Response{}, err
	}

	r.Request = rest.RequestInfo{Method: httpResp.Request.Method, URL: httpResp.Request.URL.String()}
	r.StatusCode = httpResp.StatusCode
	r.Data = body

	if !rest.IsSuccess(httpResp) {
		return Response{}, api.APIError{
			StatusCode: httpResp.StatusCode,
			Body:       body,
			Request:    rest.RequestInfo{Method: httpResp.Request.Method, URL: httpResp.Request.URL.String()},
		}
	}
	contentType := httpResp.Header["Content-Type"][0]
	boundaryIndex := strings.Index(contentType, "boundary=")
	if boundaryIndex == -1 {
		return r, fmt.Errorf("no boundary parameter found in Content-Type header")
	}
	boundary := contentType[boundaryIndex+len("boundary="):]

	reader := multipart.NewReader(httpResp.Body, boundary)

	form, err := reader.ReadForm(0)
	if err != nil {
		return r, fmt.Errorf("unable to read multipart form: %w", err)
	}

	if len(form.Value["metadata"]) == 0 {
		return r, fmt.Errorf("metadata field not found in response")
	}

	err = json.Unmarshal([]byte(form.Value["metadata"][0]), &r)
	if err != nil {
		return r, fmt.Errorf("unable to unmarshal metadata: %w", err)
	}

	file, err := form.File["content"][0].Open()
	if err != nil {
		return r, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	fileContent := new(bytes.Buffer)
	_, err = fileContent.ReadFrom(file)
	if err != nil {
		return r, fmt.Errorf("unable to read file: %w", err)
	}
	r.Data = fileContent.Bytes()

	return r, nil
}

func (c Client) List(ctx context.Context, filter string) (ListResponse, error) {
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
			queryParams["page-key"] = []string{*result.NextPageKey}
		}

		ro := rest.RequestOptions{
			QueryParams: queryParams,
		}

		resp, err := c.client.List(ctx, ro)
		if err != nil {
			return ListResponse{}, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
			return ListResponse{}, api.NewAPIErrorFromResponseAndBody(resp, body)
		}
		if !rest.IsSuccess(resp) {
			return ListResponse{},
				api.APIError{
					StatusCode: resp.StatusCode,
					Body:       body,
					Request:    rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()},
				}

		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			return ListResponse{}, err
		}

		for i := range result.Documents {
			result.Documents[i].Request = rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()}
			result.Documents[i].StatusCode = resp.StatusCode
		}

		retVal.Responses = append(retVal.Responses, result.Documents...)
		retVal.StatusCode = resp.StatusCode
	}

	return retVal, nil
}

func (c Client) Create(ctx context.Context, name string, isPrivate bool, externalId string, data []byte, documentType DocumentType) (api.Response, error) {
	d := documents.Document{
		Kind:       string(documentType),
		Name:       name,
		Public:     !isPrivate,
		ExternalID: externalId,
		Content:    data,
	}

	resp, err := c.create(ctx, d)
	if err != nil {
		return api.Response{}, err
	}

	var md documentMetaData
	if md, err = metadata(resp.Data); err != nil {
		return api.Response{}, err
	}

	r, err := c.patch(ctx, md.ID, md.Version, d)
	if err != nil {
		return api.Response{}, err
	}

	return r, nil
}

func (c Client) Update(ctx context.Context, id string, name string, isPrivate bool, data []byte, documentType DocumentType) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf("id must be non-empty")
	}

	getResp, err := c.Get(ctx, id)
	if err != nil {
		return api.Response{}, err
	}

	if !(getResp.StatusCode >= 200 && getResp.StatusCode <= 299) {
		return api.Response{}, api.APIError{
			StatusCode: getResp.StatusCode,
			Body:       getResp.Data,
			Request:    rest.RequestInfo{Method: getResp.Request.Method, URL: getResp.Request.URL},
		}
	}

	d := documents.Document{
		Kind:    string(documentType),
		Name:    name,
		Public:  !isPrivate,
		Content: data,
	}

	patchResp, err := c.client.Patch(ctx, id, getResp.Version, d)
	if err != nil {
		return api.Response{}, err
	}
	defer patchResp.Body.Close()

	respBody, err := io.ReadAll(patchResp.Body)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
		return api.Response{}, api.NewAPIErrorFromResponseAndBody(patchResp, respBody)
	}

	if !rest.IsSuccess(patchResp) {
		return api.Response{}, api.NewAPIErrorFromResponseAndBody(patchResp, respBody)
	}

	return api.NewResponseFromHTTPResponseAndBody(patchResp, respBody), nil
}

func (c Client) Delete(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf("id must be non-empty")
	}

	getResp, err := c.Get(ctx, id)
	if err != nil {
		return api.Response{}, err
	}

	resp, err := c.client.Delete(ctx, id, getResp.Version)
	if err != nil {
		return api.Response{}, err
	}

	if !rest.IsSuccess(resp) {
		return api.Response{}, api.NewAPIErrorFromResponseAndBody(resp, nil)
	}

	resp, err = c.client.Trash(ctx, id)
	if err != nil {
		return api.Response{}, err
	}

	if !rest.IsSuccess(resp) {
		return api.Response{}, api.NewAPIErrorFromResponseAndBody(resp, nil)
	}

	return api.NewResponseFromHTTPResponseAndBody(resp, nil), nil
}

func (c Client) create(ctx context.Context, d documents.Document) (api.Response, error) {
	resp, err := c.client.Create(ctx, d)
	if err != nil {
		return api.Response{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
		return api.Response{}, api.NewAPIErrorFromResponseAndBody(resp, body)
	}

	if !rest.IsSuccess(resp) {
		return api.Response{}, api.NewAPIErrorFromResponseAndBody(resp, body)
	}

	return api.NewResponseFromHTTPResponseAndBody(resp, body), nil
}

func (c Client) patch(ctx context.Context, id string, version int, d documents.Document) (api.Response, error) {
	resp, err := c.client.Patch(ctx, id, version, d)
	if err != nil {
		return api.Response{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
		return api.Response{}, err
	}

	if !rest.IsSuccess(resp) {
		return api.Response{}, api.NewAPIErrorFromResponseAndBody(resp, body)
	}

	var metadata map[string]any
	if err := json.Unmarshal(body, &metadata); err != nil {
		return api.Response{}, err
	}
	var tmp []byte
	if tmp, err = json.Marshal(metadata["documentMetadata"]); err != nil {
		return api.Response{}, err
	}
	body = tmp

	return api.NewResponseFromHTTPResponseAndBody(resp, body), nil
}

func (c Client) get(ctx context.Context, id string) (api.MultipartResponse, error) {
	resp, err := c.client.Get(ctx, id)
	if err != nil {
		return api.MultipartResponse{}, fmt.Errorf("failed to get document resource with id %s: %w", id, err)
	}

	boundary, err := extractBoundary(resp)
	if err != nil {
		return api.MultipartResponse{}, fmt.Errorf("failed to read the content of the document resource with id %s: %w", id, err)
	}

	var parts []api.Part

	r := multipart.NewReader(resp.Body, boundary)
	for p, err := r.NextPart(); err != io.EOF; {
		if err != nil {
			return api.MultipartResponse{}, fmt.Errorf("failed to read the content of the document resource with id %s: %w", id, err)
		}
		buf, err := io.ReadAll(p)
		if err != nil {
			return api.MultipartResponse{}, fmt.Errorf("failed to read the content of the document resource with id %s: %w", id, err)
		}

		parts = append(parts, api.Part{
			FormName: p.FormName(),
			FileName: p.FileName(),
			Content:  buf,
		})
	}

	return *api.NewMultipartResponse(resp, parts...), nil
}

func extractBoundary(resp *http.Response) (string, error) {
	t, ps, err := mime.ParseMediaType(resp.Header.Get("content-type"))
	if t != "multipart/x-mixed-replace" {
		return "", errors.New("http response is not multipart/x-mixed-replace")
	}
	if err != nil {
		return "", err
	}
	return ps["boundary"], nil
}
