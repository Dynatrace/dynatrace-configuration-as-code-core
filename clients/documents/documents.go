package documents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/clients/documents"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/go-logr/logr"
	"io"
	"mime/multipart"
	"net/url"
	"strconv"
	"strings"
)

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

const bodyReadErrMsg = "unable to read API response body"

const optimisticLockingHeader = "optimistic-locking-version"

type DocumentType string

const (
	Dashboard DocumentType = "dashboard"
	Notebook  DocumentType = "notebook"
)

// Client is the HTTP client to be used for interacting with the Document API
type Client struct {
	client *documents.Client
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

// ListResponse is a list of API Responses
type ListResponse struct {
	api.Response
	Responses []Response
}

// NewClient creates a new document client
func NewClient(client *rest.Client) *Client {
	c := &Client{client: documents.NewClient(client)}
	return c
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

func (c Client) Create(ctx context.Context, name string, isPrivate bool, externalId string, data []byte, documentType DocumentType) (Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("type", string(documentType)); err != nil {
		return Response{}, err
	}
	if err := writer.WriteField("name", name); err != nil {
		return Response{}, err
	}

	if err := writer.WriteField("isPrivate", strconv.FormatBool(isPrivate)); err != nil {
		return Response{}, err
	}

	if externalId != "" {
		if err := writer.WriteField("externalId", externalId); err != nil {
			return Response{}, err
		}
	}

	part, err := writer.CreatePart(map[string][]string{
		"Content-Type":        {"application/json"},
		"Content-Disposition": {fmt.Sprintf(`form-data; name="content"; filename="%s"`, name)},
	})
	if err != nil {
		return Response{}, err
	}

	if _, err = part.Write(data); err != nil {
		return Response{}, err
	}
	if err = writer.Close(); err != nil {
		return Response{}, err
	}

	resp, err := c.client.Create(ctx, body.Bytes(), rest.RequestOptions{
		ContentType: writer.FormDataContentType(),
	})
	if err != nil {
		return Response{}, err
	}

	if !rest.IsSuccess(resp) {
		return Response{}, api.APIError{
			StatusCode: resp.StatusCode,
			Body:       body.Bytes(),
			Request:    rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()},
		}
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
		return Response{}, api.NewAPIErrorFromResponseAndBody(resp, respBody)
	}

	if err = resp.Body.Close(); err != nil {
		return Response{}, err
	}

	r := Response{
		Response: api.Response{
			StatusCode: resp.StatusCode,
			Data:       respBody,
			Request:    rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()},
		},
	}

	if err = json.Unmarshal(respBody, &r); err != nil {
		return r, err
	}

	return r, nil
}

func (c Client) Update(ctx context.Context, id string, name string, isPrivate bool, data []byte, documentType DocumentType) (Response, error) {
	if id == "" {
		return Response{}, fmt.Errorf("id must be non-empty")
	}

	getResp, err := c.Get(ctx, id)
	if err != nil {
		return Response{}, err
	}

	if !(getResp.StatusCode >= 200 && getResp.StatusCode <= 299) {
		return Response{}, api.APIError{
			StatusCode: getResp.StatusCode,
			Body:       getResp.Data,
			Request:    rest.RequestInfo{Method: getResp.Request.Method, URL: getResp.Request.URL},
		}
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err = writer.WriteField("type", string(documentType)); err != nil {
		return getResp, err
	}
	if err = writer.WriteField("name", name); err != nil {
		return getResp, err
	}

	if err := writer.WriteField("isPrivate", strconv.FormatBool(isPrivate)); err != nil {
		return Response{}, err
	}

	part, err := writer.CreatePart(map[string][]string{
		"Content-Type":        {"application/json"},
		"Content-Disposition": {fmt.Sprintf(`form-data; name="content"; filename="%s"`, name)},
	})
	if err != nil {
		return getResp, err
	}

	if _, err = part.Write(data); err != nil {
		return Response{}, err
	}
	if err = writer.Close(); err != nil {
		return Response{}, err
	}

	patchResp, err := c.client.Patch(ctx, id, body.Bytes(), rest.RequestOptions{
		QueryParams: map[string][]string{optimisticLockingHeader: {fmt.Sprint(getResp.Version)}},
		ContentType: writer.FormDataContentType(),
	})

	if err != nil {
		return Response{}, err
	}

	respBody, err := io.ReadAll(patchResp.Body)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, bodyReadErrMsg)
		return Response{}, api.NewAPIErrorFromResponseAndBody(patchResp, respBody)
	}
	if err = patchResp.Body.Close(); err != nil {
		return Response{}, err
	}

	if !rest.IsSuccess(patchResp) {
		return Response{}, api.APIError{
			StatusCode: patchResp.StatusCode,
			Body:       respBody,
			Request:    rest.RequestInfo{Method: patchResp.Request.Method, URL: patchResp.Request.URL.String()},
		}
	}

	type documentMetaData struct {
		ID         string `json:"id"`
		ExternalID string `json:"externalId"`
		Actor      string `json:"actor"`
		Owner      string `json:"owner"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		Version    int    `json:"version"`
	}
	type metadata struct {
		DocumentMetaData documentMetaData `json:"documentMetadata"`
	}

	var r metadata
	if err = json.Unmarshal(respBody, &r); err != nil {
		return Response{}, err
	}

	return Response{
		ID:      r.DocumentMetaData.ID,
		Actor:   r.DocumentMetaData.Actor,
		Owner:   r.DocumentMetaData.Owner,
		Name:    r.DocumentMetaData.Name,
		Type:    r.DocumentMetaData.Type,
		Version: r.DocumentMetaData.Version,

		Response: api.Response{
			StatusCode: patchResp.StatusCode,
			Data:       respBody,
			Request:    rest.RequestInfo{Method: patchResp.Request.Method, URL: patchResp.Request.URL.String()},
		},
	}, nil

}

func (c Client) Delete(ctx context.Context, id string) (Response, error) {
	if id == "" {
		return Response{}, fmt.Errorf("id must be non-empty")
	}

	getResp, err := c.Get(ctx, id)
	if err != nil {
		return Response{}, err
	}

	if !(getResp.StatusCode >= 200 && getResp.StatusCode <= 299) {
		return Response{}, api.APIError{
			StatusCode: getResp.StatusCode,
			Body:       getResp.Data,
			Request:    rest.RequestInfo{Method: getResp.Request.Method, URL: getResp.Request.URL},
		}
	}

	resp, err := c.client.Delete(ctx, id, rest.RequestOptions{
		QueryParams: map[string][]string{optimisticLockingHeader: {fmt.Sprint(getResp.Version)}},
	})
	if err != nil {
		return Response{}, err
	}

	if !rest.IsSuccess(resp) {
		return Response{}, api.APIError{
			StatusCode: resp.StatusCode,
			Request:    rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()},
		}
	}

	resp, err = c.client.Trash(ctx, id)
	if err != nil {
		return Response{}, err
	}

	if !rest.IsSuccess(resp) {
		return Response{}, api.APIError{
			StatusCode: resp.StatusCode,
			Request:    rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()},
		}
	}

	return Response{
		Response: api.Response{
			StatusCode: resp.StatusCode,
			Data:       nil,
			Request:    rest.RequestInfo{Method: resp.Request.Method, URL: resp.Request.URL.String()},
		},
	}, nil
}
