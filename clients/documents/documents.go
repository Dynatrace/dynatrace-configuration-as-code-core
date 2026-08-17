/*
 * @license
 * Copyright 2026 Dynatrace LLC
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
	"log/slog"
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

	documentsResource = "documents"
	trashResource     = "documents-trash"
)

var (
	idValidationErr = api.ValidationError{Resource: documentsResource, Field: "id", Reason: "is empty"}
	errNoMetadata   = errors.New("metadata field not found in response")
	errNoContent    = errors.New("content field not found in response")
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
		return Response{}, idValidationErr
	}

	return c.get(ctx, id, true)
}

func readMetadata(id string, form *multipart.Form) (Metadata, error) {
	if len(form.Value["metadata"]) == 0 {
		return Metadata{}, api.RuntimeError{Resource: documentsResource, Identifier: id, Wrapped: errNoMetadata}
	}

	md, err := UnmarshallMetadata([]byte(form.Value["metadata"][0]))
	if err != nil {
		return Metadata{}, api.RuntimeError{Resource: documentsResource, Identifier: id, Reason: "unmarshalling metadata failed", Wrapped: err}
	}
	return md, nil
}

func readFileContent(id string, form *multipart.Form) ([]byte, error) {
	if len(form.File["content"]) == 0 {
		return nil, api.RuntimeError{Resource: documentsResource, Identifier: id, Wrapped: errNoContent}
	}

	file, err := form.File["content"][0].Open()
	if err != nil {
		return nil, api.RuntimeError{Resource: documentsResource, Identifier: id, Reason: "unable to open file", Wrapped: err}
	}
	defer file.Close()

	fileContent := new(bytes.Buffer)
	_, err = fileContent.ReadFrom(file)
	if err != nil {
		return nil, api.RuntimeError{Resource: documentsResource, Identifier: id, Reason: "unable to read file", Wrapped: err}
	}
	return fileContent.Bytes(), nil
}

// listAddFields are the fields the list endpoint omits from its reduced metadata
var listAddFields = []string{"description", "labels", "originExtensionId"}

// List returns every document matching filter. The list endpoint returns a
// reduced metadata object
func (c Client) List(ctx context.Context, filter string) (ListResponse, error) {
	type listResponse struct {
		Documents   []Metadata `json:"documents"`
		NextPageKey *string    `json:"nextPageKey"`
	}

	var retVal ListResponse
	nextPageKey := new("")

	for nextPageKey != nil {

		queryParams := url.Values{"filter": {filter}, "add-fields": listAddFields}
		if *nextPageKey != "" {
			queryParams["page-key"] = []string{*nextPageKey}
		}

		ro := rest.RequestOptions{QueryParams: queryParams}

		resp, err := c.restClient.GET(ctx, documentResourcePath, ro)
		if err != nil {
			return ListResponse{}, api.ClientError{Resource: documentsResource, Operation: http.MethodGet, Wrapped: err}
		}

		res, err := api.NewResponseFromHTTPResponse(resp)
		if err != nil {
			return ListResponse{}, api.ClientError{Resource: documentsResource, Operation: http.MethodGet, Wrapped: err}
		}

		var result listResponse
		if err := json.Unmarshal(res.Data, &result); err != nil {
			return ListResponse{}, api.RuntimeError{Resource: documentsResource, Reason: "unmarshalling failed", Wrapped: err}
		}
		nextPageKey = result.NextPageKey

		for _, doc := range result.Documents {
			retVal.Responses = append(retVal.Responses, Response{
				Response: api.Response{
					Request:    rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()},
					StatusCode: resp.StatusCode,
				},
				Metadata: doc,
			})
		}

		retVal.StatusCode = resp.StatusCode
	}

	return retVal, nil
}

// Create creates a new document from meta and content. Only the settable
// Metadata fields are sent (Type, Name, IsPrivate, ID, Description, Labels,
// IsReshareable); leave ID empty to let the API generate one. Server-managed
// fields on meta are ignored (see writeDocument).
func (c Client) Create(ctx context.Context, meta Metadata, content []byte) (api.Response, error) {
	body := &bytes.Buffer{}
	contentType, _, err := writeDocument(body, meta, content)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: documentsResource, Reason: "failed to write document body", Wrapped: err}
	}

	httpResp, err := c.restClient.POST(ctx, documentResourcePath, body, rest.RequestOptions{
		ContentType: contentType,
	})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: documentsResource, Operation: http.MethodPost, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: documentsResource, Operation: http.MethodPost, Wrapped: err}
	}

	var md Metadata
	if md, err = UnmarshallMetadata(resp.Data); err != nil {
		return api.Response{}, api.RuntimeError{Resource: documentsResource, Reason: "unmarshalling failed", Wrapped: err}
	}

	r, err := c.patchWithRetry(ctx, md.ID, md.Version, meta, content)
	if err != nil {
		if !api.IsNotFoundError(err) {
			if err1 := c.deleteAndTrash(ctx, md.ID); err1 != nil {
				return api.Response{}, errors.Join(err, err1)
			}
		}
		return api.Response{}, err
	}
	return r, nil
}

// Update replaces the document identified by meta.ID with meta and content. Only
// the settable Metadata fields are sent (Type, Name, IsPrivate, ID, Description,
// Labels, IsReshareable). The current server version is fetched and used as the
// optimistic-locking version (last-write-wins); a caller-supplied meta.Version
// is ignored.
func (c Client) Update(ctx context.Context, meta Metadata, content []byte) (api.Response, error) {
	if meta.ID == "" {
		return api.Response{}, idValidationErr
	}

	resp, err := c.get(ctx, meta.ID, false)
	if err != nil {
		return api.Response{}, err
	}

	return c.patch(ctx, meta.ID, resp.Version, meta, content)
}

func (c Client) Delete(ctx context.Context, id string) (api.Response, error) {
	if id == "" {
		return api.Response{}, idValidationErr
	}

	return api.Response{}, c.deleteAndTrash(ctx, id)
}

func (c Client) patchWithRetry(ctx context.Context, id string, version int, meta Metadata, content []byte) (resp api.Response, err error) {
	const maxRetries = 5
	const retryDelay = 200 * time.Millisecond
	for range maxRetries {
		if resp, err = c.patch(ctx, id, version, meta, content); api.IsNotFoundError(err) {
			time.Sleep(retryDelay)
			continue
		}
		break
	}
	return
}

func (c Client) patch(ctx context.Context, id string, version int, meta Metadata, content []byte) (api.Response, error) {
	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: documentsResource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	body := &bytes.Buffer{}
	contentType, _, err := writeDocument(body, meta, content)
	if err != nil {
		return api.Response{}, api.RuntimeError{Resource: documentsResource, Identifier: id, Reason: "failed to write document body", Wrapped: err}
	}

	httpResp, err := c.restClient.PATCH(ctx, path, body, rest.RequestOptions{
		ContentType: contentType,
		QueryParams: url.Values{optimisticLockingHeader: []string{strconv.Itoa(version)}},
	})
	if err != nil {
		return api.Response{}, api.ClientError{Resource: documentsResource, Identifier: id, Operation: http.MethodPatch, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return api.Response{}, api.ClientError{Resource: documentsResource, Identifier: id, Operation: http.MethodPatch, Wrapped: err}
	}

	tmp, err := extractMetadata(resp.Data)
	if err != nil {
		return resp, api.RuntimeError{Resource: documentsResource, Identifier: id, Reason: "extracting metadata failed", Wrapped: err}
	}
	resp.Data = tmp

	return resp, nil
}

func (c Client) get(ctx context.Context, id string, readContent bool) (Response, error) {
	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return Response{}, api.RuntimeError{Resource: documentsResource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	httpResp, err := c.restClient.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return Response{}, api.ClientError{Resource: documentsResource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}

	resp, err := api.NewResponseFromHTTPResponse(httpResp)
	if err != nil {
		return Response{}, api.ClientError{Resource: documentsResource, Identifier: id, Operation: http.MethodGet, Wrapped: err}
	}

	boundary, err := extractBoundary(resp)
	if err != nil {
		return Response{}, api.RuntimeError{Resource: documentsResource, Identifier: id, Reason: "failed to read response content type", Wrapped: err}
	}

	reader := multipart.NewReader(bytes.NewReader(resp.Data), boundary)

	form, err := reader.ReadForm(0)
	if err != nil {
		return Response{}, api.RuntimeError{Resource: documentsResource, Identifier: id, Reason: "unable to read multipart form", Wrapped: err}
	}
	defer func() {
		err := form.RemoveAll()
		if err != nil {
			slog.WarnContext(ctx, "Failed to remove multipart form temporary files", slog.String("error", err.Error()))
		}
	}()

	metadata, err := readMetadata(id, form)
	if err != nil {
		return Response{}, err
	}

	if readContent {
		fileContent, err := readFileContent(id, form)
		if err != nil {
			return Response{}, err
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

func (c Client) deleteAndTrash(ctx context.Context, id string) error {
	if err := c.delete(ctx, id); err != nil && !api.IsNotFoundError(err) {
		return err
	}
	return c.trash(ctx, id)
}

func (c Client) delete(ctx context.Context, id string) error {
	resp, err := c.get(ctx, id, false)
	if err != nil {
		return err
	}

	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return api.RuntimeError{Resource: documentsResource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	r, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{
		QueryParams: map[string][]string{optimisticLockingHeader: {strconv.Itoa(resp.Version)}},
	})
	if err != nil {
		return api.ClientError{Resource: documentsResource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}

	_, err = api.NewResponseFromHTTPResponse(r)
	if err != nil {
		return api.ClientError{Resource: documentsResource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}

	return nil
}

func (c Client) trash(ctx context.Context, id string) error {
	path, err := url.JoinPath(trashResourcePath, id)
	if err != nil {
		return api.RuntimeError{Resource: trashResource, Identifier: id, Reason: "failed to construct URL", Wrapped: err}
	}

	resp, err := c.restClient.DELETE(ctx, path, rest.RequestOptions{})
	if err != nil {
		return api.ClientError{Resource: trashResource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
	}

	_, err = api.NewResponseFromHTTPResponse(resp)
	if err != nil {
		return api.ClientError{Resource: trashResource, Identifier: id, Operation: http.MethodDelete, Wrapped: err}
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
