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
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	bodyReadErrMsg          = "unable to read API response body"
	documentResourcePath    = "/platform/document/v1/documents"
	optimisticLockingHeader = "optimistic-locking-version"
	trashResourcePath       = "/platform/document/v1/trash/documents"
)

// DocumentType defines the *known* types of documents. It is possible to pass an arbitrary string in consumers
// to download any kind of document.
type DocumentType = string

const (
	Dashboard DocumentType = "dashboard"
	Notebook  DocumentType = "notebook"
	Launchpad DocumentType = "launchpad"
)

// Client is the HTTP client to be used for interacting with the Document API
type Client struct {
	restClient *rest.Client
}

// NewClient creates a new document client
func NewClient(client *rest.Client) *Client {
	c := &Client{restClient: client}
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

func (c Client) Get(ctx context.Context, id string) (Response, error) {
	var r Response

	httpResp, err := c.get(ctx, id)
	if err != nil {
		return Response{}, fmt.Errorf("failed to getAsMultipart document resource with id %s: %w", id, err)
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

		ro := rest.RequestOptions{QueryParams: queryParams}

		resp, err := c.restClient.GET(ctx, documentResourcePath, ro)
		if err != nil {
			return ListResponse{}, fmt.Errorf("unable to get objects: %w", err)
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
	d := Document{
		Kind:       documentType,
		Name:       name,
		Public:     !isPrivate,
		ExternalID: externalId,
		Content:    data,
	}

	body := &bytes.Buffer{}
	writer, err := d.write(body)
	if err != nil {
		return api.Response{}, err
	}

	httpResp, err := c.restClient.POST(ctx, documentResourcePath, body, rest.RequestOptions{
		ContentType: writer.FormDataContentType(),
	})
	resp, err := processHttpResponse(httpResp, err)
	if err != nil {
		return api.Response{}, fmt.Errorf("unable to create object: %w", err)
	}

	var md Metadata
	if md, err = UnmarshallMetadata(resp.Data); err != nil {
		return api.Response{}, err
	}

	r, err := c.patchWithRetry(ctx, md.ID, md.Version, d)
	if err != nil {
		if !api.IsNotFoundError(err) {
			if _, err1 := c.delete(ctx, md.ID, md.Version); err1 != nil {
				return api.Response{}, errors.Join(err, err1)
			}
		}
		return api.Response{}, err
	}
	return r, nil
}

func (c Client) Update(ctx context.Context, id string, name string, isPrivate bool, data []byte, documentType DocumentType) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf("id must be non-empty")
	}

	resp, err := c.getAsMultipart(ctx, id)
	if !resp.IsSuccess() {
		return api.Response{}, err
	}

	part, _ := resp.GetPartWithName("metadata")
	md, err := UnmarshallMetadata(part.Content)
	if err != nil {
		return api.Response{}, err
	}

	d := Document{
		Kind:    documentType,
		Name:    name,
		Public:  !isPrivate,
		Content: data,
	}

	return c.patch(ctx, id, md.Version, d)
}

func (c Client) Delete(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf("id must be non-empty")
	}

	resp, err := c.getAsMultipart(ctx, id)
	if !resp.IsSuccess() {
		return api.Response{}, err
	}

	part, _ := resp.GetPartWithName("metadata")
	md, err := UnmarshallMetadata(part.Content)
	if err != nil {
		return api.Response{}, err
	}

	return c.delete(ctx, id, md.Version)
}

func (c Client) patchWithRetry(ctx context.Context, id string, version int, d Document) (resp api.Response, err error) {
	const maxRetries = 5
	const retryDelay = 200 * time.Millisecond
	for r := 0; r < maxRetries; r++ {
		if resp, err = c.patch(ctx, id, version, d); api.IsNotFoundError(err) {
			time.Sleep(retryDelay)
			continue
		}
		break
	}
	return
}

func (c Client) patch(ctx context.Context, id string, version int, d Document) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf("id is missing")
	}
	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf("failed to create URL: %w", err)
	}

	body := &bytes.Buffer{}
	writer, err := d.write(body)
	if err != nil {
		return api.Response{}, err
	}

	httpResp, err := c.restClient.PATCH(ctx, path, body, rest.RequestOptions{
		ContentType: writer.FormDataContentType(),
		QueryParams: url.Values{optimisticLockingHeader: []string{strconv.Itoa(version)}},
	})
	resp, err := processHttpResponse(httpResp, err)
	if err != nil {
		return api.Response{}, fmt.Errorf("unable to update object: %w", err)
	}

	tmp, err := extractMetadata(resp.Data)
	if err != nil {
		return resp, fmt.Errorf("extracting metadata failed: %w", err)
	}
	resp.Data = tmp

	return resp, nil
}

func (c Client) get(ctx context.Context, id string) (*http.Response, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}

	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := c.restClient.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get object with ID %s: %w", id, err)
	}

	return resp, err
}

func (c Client) getAsMultipart(ctx context.Context, id string) (api.MultipartResponse, error) {
	resp, err := c.get(ctx, id)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return api.MultipartResponse{}, fmt.Errorf("failed to getAsMultipart document resource with id %s: %w", id, err)
	}

	if !rest.IsSuccess(resp) {
		return api.MultipartResponse{}, api.NewAPIErrorFromResponse(resp)
	}

	boundary, err := extractBoundary(resp)
	if err != nil {
		return api.MultipartResponse{}, fmt.Errorf("failed to read the content of the document resource with id %s: %w", id, err)
	}

	var parts []api.Part

	r := multipart.NewReader(resp.Body, boundary)
	for p, err := r.NextPart(); err != io.EOF; p, err = r.NextPart() {
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

	out := *api.NewMultipartResponse(resp, parts...)

	if _, ok := out.GetPartWithName("metadata"); !ok {
		return out, fmt.Errorf("metadata not present for object with id %s", id)
	}
	if _, ok := out.GetPartWithName("content"); !ok {
		return out, fmt.Errorf("content not present for object with id %s", id)
	}

	return *api.NewMultipartResponse(resp, parts...), nil
}
func extractBoundary(resp *http.Response) (string, error) {
	t, ps, err := mime.ParseMediaType(resp.Header.Get("content-type"))
	if !strings.HasPrefix(t, "multipart") {
		return "", errors.New("http response is not multipart")
	}
	if err != nil {
		return "", err
	}
	return ps["boundary"], nil
}

func (c Client) delete(ctx context.Context, id string, version int) (api.Response, error) {
	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{
		QueryParams:           map[string][]string{optimisticLockingHeader: {strconv.Itoa(version)}},
		CustomShouldRetryFunc: rest.RetryOnFailureExcept404,
	})
	r, err := processHttpResponse(resp, err)
	if err != nil {
		return r, err
	}

	return processHttpResponse(c.trash(ctx, id))
}

func (c Client) trash(ctx context.Context, id string) (*http.Response, error) {
	path, err := url.JoinPath(trashResourcePath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to trash object: %w", err)
	}
	return resp, err
}

func extractMetadata(in []byte) (out []byte, err error) {
	var metadata map[string]any
	if err = json.Unmarshal(in, &metadata); err != nil {
		return
	}
	return json.Marshal(metadata["documentMetadata"])
}

func processHttpResponse(resp *http.Response, err error) (api.Response, error) {
	if err != nil {
		return api.Response{}, err
	}

	return api.NewResponseFromHTTPResponse(resp)
}
