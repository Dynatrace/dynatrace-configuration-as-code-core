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
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	documentResourcePath    = "/platform/document/v1/documents"
	trashResourcePath       = "/platform/document/v1/trash/documents"
	optimisticLockingHeader = "optimistic-locking-version"

	errMsg         = "failed to %s document: %w"
	errMsgWithName = "failed to %s document with name %s: %w"
	errMsgWithID   = "failed to %s document with ID %s: %w"

	getOperation    = "get"
	listOperation   = "list"
	createOperation = "create"
	deleteOperation = "delete"
	trashOperation  = "trash"
	updateOperation = "update"
)

var (
	ErrIDEmpty    = fmt.Errorf("id must be non-empty")
	ErrNoMetadata = fmt.Errorf("metadata field not found in response")
	ErrNoContent  = fmt.Errorf("content field not found in response")
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
	if id == "" {
		return Response{}, fmt.Errorf(errMsg, getOperation, ErrIDEmpty)
	}

	return c.get(ctx, id, true)
}

func readMetadata(form *multipart.Form) (Metadata, error) {
	if len(form.Value["metadata"]) == 0 {
		return Metadata{}, ErrNoMetadata
	}

	return UnmarshallMetadata([]byte(form.Value["metadata"][0]))
}

func readFileContent(form *multipart.Form) ([]byte, error) {
	if len(form.File["content"]) == 0 {
		return nil, ErrNoContent
	}
	file, err := form.File["content"][0].Open()
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	fileContent := new(bytes.Buffer)
	_, err = fileContent.ReadFrom(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}
	return fileContent.Bytes(), nil
}

func (c Client) List(ctx context.Context, filter string) (ListResponse, error) {
	type listResponse struct {
		TotalCount  int        `json:"totalCount"`
		Documents   []Metadata `json:"documents"`
		NextPageKey *string    `json:"nextPageKey"`
	}

	var retVal ListResponse
	var result listResponse
	var initialPage = ""
	result.NextPageKey = &initialPage

	for result.NextPageKey != nil {

		queryParams := url.Values{"filter": {filter}, "add-field": {"originExtensionId"}}
		if *result.NextPageKey != "" {
			queryParams["page-key"] = []string{*result.NextPageKey}
		}

		ro := rest.RequestOptions{QueryParams: queryParams}

		resp, err := c.restClient.GET(ctx, documentResourcePath, ro)
		if err != nil {
			return ListResponse{}, fmt.Errorf(errMsg, listOperation, err)
		}
		res, err := api.NewResponseFromHTTPResponse(resp)

		if err != nil {
			return ListResponse{}, fmt.Errorf(errMsg, listOperation, err)
		}

		err = json.Unmarshal(res.Data, &result)
		if err != nil {
			return ListResponse{}, err
		}

		for _, metadata := range result.Documents {
			retVal.Responses = append(retVal.Responses, Response{
				Response: api.Response{
					Request:    rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()},
					StatusCode: resp.StatusCode,
				},
				Metadata: metadata,
			})
		}

		retVal.StatusCode = resp.StatusCode
	}

	return retVal, nil
}

func (c Client) Create(ctx context.Context, name string, isPrivate bool, id string, data []byte, documentType DocumentType) (api.Response, error) {
	d := Document{
		Kind:    documentType,
		Name:    name,
		Public:  !isPrivate,
		ID:      id,
		Content: data,
	}

	body := &bytes.Buffer{}
	writer, err := d.write(body)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, createOperation, name, err)
	}

	httpResp, err := c.restClient.POST(ctx, documentResourcePath, body, rest.RequestOptions{
		ContentType: writer.FormDataContentType(),
	})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, createOperation, name, err)
	}
	resp, err := api.NewResponseFromHTTPResponse(httpResp)

	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, createOperation, name, err)
	}

	var md Metadata
	if md, err = UnmarshallMetadata(resp.Data); err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithName, createOperation, name, err)
	}

	r, err := c.patchWithRetry(ctx, md.ID, md.Version, d)
	if err != nil {
		if !api.IsNotFoundError(err) {
			if _, err1 := c.delete(ctx, md.ID, md.Version); err1 != nil {
				return api.Response{}, fmt.Errorf(errMsgWithName, createOperation, name, errors.Join(err, err1))
			}
		}
		return api.Response{}, fmt.Errorf(errMsgWithName, createOperation, name, err)
	}
	return r, nil
}

func (c Client) Update(ctx context.Context, id string, name string, isPrivate bool, data []byte, documentType DocumentType) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf(errMsg, updateOperation, ErrIDEmpty)
	}

	resp, err := c.get(ctx, id, false)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsg, updateOperation, err)
	}

	d := Document{
		Kind:    documentType,
		Name:    name,
		Public:  !isPrivate,
		Content: data,
	}

	return c.patch(ctx, id, resp.Version, d)
}

func (c Client) Delete(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, fmt.Errorf(errMsg, deleteOperation, ErrIDEmpty)
	}

	resp, err := c.get(ctx, id, false)

	if err != nil {
		return api.Response{}, err
	}

	return c.delete(ctx, id, resp.Version)
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
	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, updateOperation, id, err)
	}

	body := &bytes.Buffer{}
	writer, err := d.write(body)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, updateOperation, id, err)
	}

	httpResp, err := c.restClient.PATCH(ctx, path, body, rest.RequestOptions{
		ContentType: writer.FormDataContentType(),
		QueryParams: url.Values{optimisticLockingHeader: []string{strconv.Itoa(version)}},
	})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, updateOperation, id, err)
	}
	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, updateOperation, id, err)
	}

	tmp, err := extractMetadata(resp.Data)
	if err != nil {
		return resp, fmt.Errorf(errMsgWithID, updateOperation, id, fmt.Errorf("extracting metadata failed: %w", err))
	}
	resp.Data = tmp

	return resp, nil
}

func (c Client) get(ctx context.Context, id string, readContent bool) (Response, error) {
	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return Response{}, fmt.Errorf(errMsg, getOperation, err)
	}

	httpResp, err := c.restClient.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return Response{}, fmt.Errorf(errMsgWithID, getOperation, id, err)
	}
	resp, err := api.NewResponseFromHTTPResponse(httpResp)

	if err != nil {
		return Response{}, fmt.Errorf(errMsgWithID, getOperation, id, err)
	}

	boundary, err := extractBoundary(resp)
	if err != nil {
		return Response{}, fmt.Errorf(errMsgWithID, getOperation, id, err)
	}

	reader := multipart.NewReader(bytes.NewReader(resp.Data), boundary)

	form, err := reader.ReadForm(0)
	if err != nil {
		return Response{}, fmt.Errorf(errMsgWithID, getOperation, id, fmt.Errorf("unable to read multipart form: %w", err))
	}

	metadata, err := readMetadata(form)
	if err != nil {
		return Response{}, fmt.Errorf(errMsgWithID, getOperation, id, err)
	}

	if readContent {
		fileContent, err := readFileContent(form)
		if err != nil {
			return Response{}, fmt.Errorf(errMsgWithID, getOperation, id, err)
		}
		resp.Data = fileContent
	}

	return Response{
		Response: resp,
		Metadata: metadata,
	}, nil
}

func extractBoundary(resp api.Response) (string, error) {
	t, ps, err := mime.ParseMediaType(resp.Header.Get("content-type"))
	if !strings.HasPrefix(t, "multipart") {
		return "", http.ErrNotMultipart
	}
	if err != nil {
		return "", err
	}
	return ps["boundary"], nil
}

func (c Client) delete(ctx context.Context, id string, version int) (api.Response, error) {
	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, deleteOperation, id, err)
	}

	r, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{
		QueryParams:           map[string][]string{optimisticLockingHeader: {strconv.Itoa(version)}},
		CustomShouldRetryFunc: rest.RetryOnFailureExcept404,
	})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, deleteOperation, id, err)
	}
	_, err = api.NewResponseFromHTTPResponse(r)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, deleteOperation, id, err)
	}

	return c.trash(ctx, id)
}

func (c Client) trash(ctx context.Context, id string) (api.Response, error) {
	path, err := url.JoinPath(trashResourcePath, id)
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, trashOperation, id, err)
	}

	resp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.Response{}, fmt.Errorf(errMsgWithID, trashOperation, id, err)
	}
	return api.NewResponseFromHTTPResponse(resp)
}

func extractMetadata(in []byte) (out []byte, err error) {
	var metadata map[string]any
	if err = json.Unmarshal(in, &metadata); err != nil {
		return
	}
	return json.Marshal(metadata["documentMetadata"])
}
