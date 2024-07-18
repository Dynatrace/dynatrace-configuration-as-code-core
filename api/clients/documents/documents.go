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
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

const (
	documentResourcePath    = "/platform/document/v1/documents"
	optimisticLockingHeader = "optimistic-locking-version"
	trashResourcePath       = "/platform/document/v1/trash/documents"
)

type Client struct {
	client *rest.Client
}

func NewClient(client *rest.Client) *Client {
	c := &Client{
		client: client,
	}
	return c
}

func (c *Client) Get(ctx context.Context, id string) (*http.Response, error) {
	if id == "" {
		return nil, fmt.Errorf("id must be non-empty")
	}

	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := c.client.GET(ctx, path, rest.RequestOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get object with ID %s: %w", id, err)
	}

	return resp, err
}

// Create posts the content of the document to the server using http POST to create new document. NOTE: some of the arguments of the document are ignored due the design of the HTTP API.
func (c *Client) Create(ctx context.Context, doc Document) (*http.Response, error) {
	path := documentResourcePath

	body := &bytes.Buffer{}
	writer, err := doc.write(body)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.POST(ctx, path, body, rest.RequestOptions{
		ContentType: writer.FormDataContentType(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create object: %w", err)
	}

	return resp, nil
}

// Patch patches the content of the document to the server using http PATCH to change the document. NOTE: some of the arguments of the document are ignored due the design of the HTTP API.
func (c *Client) Patch(ctx context.Context, id string, version int, doc Document) (*http.Response, error) {
	if id == "" {
		return nil, fmt.Errorf("id is missing")
	}
	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	body := &bytes.Buffer{}
	writer, err := doc.write(body)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.PATCH(ctx, path, body, rest.RequestOptions{
		ContentType: writer.FormDataContentType(),
		QueryParams: url.Values{optimisticLockingHeader: []string{strconv.Itoa(version)}},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to update object: %w", err)
	}
	return resp, nil
}

func (c *Client) List(ctx context.Context, requestOptions rest.RequestOptions) (*http.Response, error) {
	path := documentResourcePath
	resp, err := c.client.GET(ctx, path, requestOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to get objects: %w", err)
	}

	return resp, err
}

func (c *Client) Update(ctx context.Context, id string, data []byte, requestOptions rest.RequestOptions) (*http.Response, error) {
	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := c.client.PUT(ctx, path, bytes.NewReader(data), requestOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to update object: %w", err)
	}
	return resp, err
}

func (c *Client) Delete(ctx context.Context, id string, requestOptions rest.RequestOptions) (*http.Response, error) {
	path, err := url.JoinPath(documentResourcePath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := c.client.DELETE(ctx, path, requestOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to delete object: %w", err)
	}
	return resp, err
}

func (c *Client) Trash(ctx context.Context, id string) (*http.Response, error) {
	path, err := url.JoinPath(trashResourcePath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	resp, err := c.client.DELETE(ctx, path, rest.RequestOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to trash object: %w", err)
	}
	return resp, err
}
