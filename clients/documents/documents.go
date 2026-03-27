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
	resource                = "documents"
)

var idValidationErr = api.ValidationError{Resource: resource, Field: "id", Reason: "is empty"}

// DocumentType defines the *known* types of documents. It is possible to pass an arbitrary string in consumers
// to download any kind of document.
type DocumentType = string

const (
	Dashboard DocumentType = "dashboard"
	Notebook  DocumentType = "notebook"
	Launchpad DocumentType = "launchpad"
)

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

// Client is used to interact with the Document API.
type Client struct {
	restClient *rest.Client
}

// NewClient creates a new document Client using the given rest.Client.
func NewClient(client *rest.Client) *Client {
	return &Client{restClient: client}
}

// Get returns one specific document by ID, including its content.
func (c Client) Get(ctx context.Context, id string) (Response, error) {
	if id == "" {
		return Response{}, idValidationErr
	}

	return c.get(ctx, id, true)
}

// List returns all documents matching the given filter.
func (c Client) List(ctx context.Context, filter string) (ListResponse, error) {
	var retVal ListResponse

	nextPageKey := ""
	for {
		var responses []Response
		var statusCode int
		var err error

		nextPageKey, responses, statusCode, err = c.listPage(ctx, filter, nextPageKey)
		if err != nil {
			return ListResponse{}, err
		}

		retVal.Responses = append(retVal.Responses, responses...)
		retVal.StatusCode = statusCode

		if nextPageKey == "" {
			break
		}
	}

	return retVal, nil
}

func (c Client) listPage(ctx context.Context, filter string, pageKey string) (string, []Response, int, error) {
	queryParams := url.Values{"filter": {filter}, "add-field": {"originExtensionId"}}
	if pageKey != "" {
		queryParams["page-key"] = []string{pageKey}
	}

	resp, err := c.restClient.GET(ctx, documentResourcePath, rest.RequestOptions{QueryParams: queryParams})
	if err != nil {
		return "", nil, 0, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	return processListResponse(resp)
}

func processListResponse(httpResponse *http.Response) (string, []Response, int, error) {
	resp, err := api.NewResponseFromHTTPResponse(httpResponse)
	if err != nil {
		return "", nil, 0, api.ClientError{Resource: resource, Operation: http.MethodGet, Wrapped: err}
	}

	var listResponse struct {
		NextPageKey *string           `json:"nextPageKey"`
		Documents   []json.RawMessage `json:"documents"`
	}

	if err := json.Unmarshal(resp.Data, &listResponse); err != nil {
		return "", nil, 0, api.RuntimeError{Resource: resource, Reason: "unmarshalling failed", Wrapped: err}
	}

	var responses []Response
	for _, doc := range listResponse.Documents {
		var metadata Metadata
		if err := json.Unmarshal(doc, &metadata); err != nil {
			return "", nil, 0, api.RuntimeError{Resource: resource, Reason: "unmarshalling failed", Wrapped: err}
		}
		responses = append(responses, Response{
			Response: api.Response{
				StatusCode: httpResponse.StatusCode,
				Request:    api.NewRequestInfoFromRequest(httpResponse.Request),
			},
			Metadata: metadata,
		})
	}

	nextPage := ""
	if listResponse.NextPageKey != nil {
		nextPage = *listResponse.NextPageKey
	}

	return nextPage, responses, httpResponse.StatusCode, nil
}

// Create creates a new document.
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
		return api.Response{}, api.RuntimeError{Resource: resource, Reason: "failed to write multipart form", Wrapped: err}
	}

	httpResp, err := c.restClient.POST(ctx, documentResourcePath, body, rest.RequestOptions{
		ContentType: writer.FormDataContentType(),
	})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPost, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Operation: http.MethodPost, Wrapped: err}
	}

	var md struct {
		ID      string `json:"id"`
		Version int    `json:"version"`
	}
	if err = json.Unmarshal(resp.Data, &md); err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Reason: "failed to unmarshal create response", Wrapped: err}
	}

	r, err := c.patchWithRetry(ctx, md.ID, md.Version, d)
	if err != nil {
		if !api.IsNotFoundError(err) {
			if err1 := c.deleteAndTrash(ctx, md.ID, md.Version); err1 != nil {
				return api.Response{}, api.ClientError{Resource: resource, Identifier: md.ID, Operation: http.MethodPost, Wrapped: errors.Join(err, err1)}
			}
		}
		return api.Response{}, api.ClientError{Resource: resource, Identifier: md.ID, Operation: http.MethodPost, Wrapped: err}
	}
	return r, nil
}

// Update updates an existing document by ID.
func (c Client) Update(ctx context.Context, id string, name string, isPrivate bool, data []byte, documentType DocumentType) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr
	}

	existing, err := c.get(ctx, id, false)
	if err != nil {
		return api.Response{}, err
	}

	version, err := getOptimisticLockingVersion(existing.Response)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to retrieve optimistic locking version", Wrapped: err}
	}

	d := Document{
		Kind:    documentType,
		Name:    name,
		Public:  !isPrivate,
		Content: data,
	}

	return c.patch(ctx, id, version, d)
}

// Delete removes a given document by ID.
func (c Client) Delete(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr
	}

	existing, err := c.get(ctx, id, false)
	if err != nil {
		return api.Response{}, err
	}

	version, err := getOptimisticLockingVersion(existing.Response)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to retrieve optimistic locking version", Wrapped: err}
	}

	if err := c.deleteAndTrash(ctx, id, version); err != nil {
		return api.Response{}, err
	}

	return api.Response{StatusCode: http.StatusOK}, nil
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
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	body := &bytes.Buffer{}
	writer, err := d.write(body)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to write multipart form", Wrapped: err}
	}

	httpResp, err := c.restClient.PATCH(ctx, path, body, rest.RequestOptions{
		ContentType: writer.FormDataContentType(),
		QueryParams: url.Values{optimisticLockingHeader: []string{strconv.Itoa(version)}},
	})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodPatch, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodPatch, Wrapped: err}
	}

	tmp, err := extractMetadata(resp.Data)
	if err != nil {
		return resp, api.RuntimeError{Resource: resource, Identifier: id, Reason: "extracting metadata failed", Wrapped: err}
	}
	resp.Data = tmp

	return resp, nil
}

func (c Client) get(ctx context.Context, id string, readContent bool) (Response, error) {
	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return Response{}, api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}

	boundary, err := extractBoundary(resp)
	if err != nil {
		return Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to extract multipart boundary", Wrapped: err}
	}

	reader := multipart.NewReader(bytes.NewReader(resp.Data), boundary)

	form, err := reader.ReadForm(0)
	if err != nil {
		return Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "unable to read multipart form", Wrapped: err}
	}

	if len(form.Value["metadata"]) == 0 {
		return Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "metadata field not found in response"}
	}

	metadataBytes := []byte(form.Value["metadata"][0])

	metadata, err := UnmarshallMetadata(metadataBytes)
	if err != nil {
		return Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to unmarshal metadata", Wrapped: err}
	}

	if readContent {
		fileContent, err := readFileContent(form)
		if err != nil {
			return Response{}, api.RuntimeError{Resource: resource, Identifier: id, Reason: "content field not found in response", Wrapped: err}
		}
		resp.Data = fileContent
	} else {
		resp.Data = metadataBytes
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

func readFileContent(form *multipart.Form) ([]byte, error) {
	if len(form.File["content"]) == 0 {
		return nil, errors.New("content field not found")
	}
	file, err := form.File["content"][0].Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileContent := new(bytes.Buffer)
	_, err = fileContent.ReadFrom(file)
	if err != nil {
		return nil, err
	}
	return fileContent.Bytes(), nil
}

func (c Client) deleteAndTrash(ctx context.Context, id string, version int) error {
	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{
		QueryParams:           url.Values{optimisticLockingHeader: {strconv.Itoa(version)}},
		CustomShouldRetryFunc: rest.RetryOnFailureExcept404,
	})
	if err != nil {
		return api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}

	_, err = api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}

	return c.trash(ctx, id)
}

func (c Client) trash(ctx context.Context, id string) error {
	path, err := url.JoinPath(trashResourcePath, id)
	if err != nil {
		return api.RuntimeError{Resource: resource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}

	_, err = api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.ClientError{Resource: resource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}
	return nil
}

func extractMetadata(in []byte) (out []byte, err error) {
	var metadata map[string]any
	if err = json.Unmarshal(in, &metadata); err != nil {
		return
	}
	return json.Marshal(metadata["documentMetadata"])
}

func getOptimisticLockingVersion(resp api.Response) (int, error) {
	var body struct {
		Version int `json:"version"`
	}

	if err := json.Unmarshal(resp.Data, &body); err != nil {
		return 0, err
	}

	return body.Version, nil
}
