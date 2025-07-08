// @license
// Copyright 2025 Dynatrace LLC
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

package buckets_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/clients/buckets"
)

type Client struct {
	get func(context.Context, string) (api.Response, error)
}

func (c Client) Get(ctx context.Context, bucketName string) (api.Response, error) {
	return c.get(ctx, bucketName)
}

func TestAwaitBucketStable_Exists(t *testing.T) {
	client := Client{get: func(context.Context, string) (api.Response, error) {
		return api.Response{
			Data: []byte(activeBucketResponse),
		}, nil
	}}
	exists, err := buckets.AwaitActiveOrNotFound(t.Context(), client, "my-bucket", time.Minute, time.Duration(0))
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestAwaitBucketStable_ExistsAfterRetry(t *testing.T) {
	apiCalls := 0
	responses := [4]string{
		deletingBucketResponse,
		creatingBucketResponse,
		updatingBucketResponse,
		activeBucketResponse,
	}
	client := Client{get: func(context.Context, string) (api.Response, error) {
		data := []byte(responses[apiCalls])
		apiCalls++
		return api.Response{
			Data: data,
		}, nil
	}}
	exists, err := buckets.AwaitActiveOrNotFound(t.Context(), client, "my-bucket", time.Minute, time.Duration(0))
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, apiCalls, 4)
}

func TestAwaitBucketStable_ReturnsOnDeadlineOfParent(t *testing.T) {
	client := Client{get: func(context.Context, string) (api.Response, error) {
		return api.Response{
			Data: []byte(activeBucketResponse),
		}, nil
	}}
	ctx, cancel := context.WithTimeout(t.Context(), time.Duration(0))
	cancel()
	exists, err := buckets.AwaitActiveOrNotFound(ctx, client, "my-bucket", time.Minute, time.Duration(0))
	assert.ErrorContains(t, err, "context canceled")
	assert.False(t, exists)
}

func TestAwaitBucketStable_ReturnsOnDeadline(t *testing.T) {
	client := Client{get: func(context.Context, string) (api.Response, error) {
		return api.Response{
			Data: []byte(activeBucketResponse),
		}, nil
	}}
	exists, err := buckets.AwaitActiveOrNotFound(t.Context(), client, "my-bucket", time.Duration(0), time.Duration(0))
	assert.ErrorContains(t, err, "context canceled")
	assert.False(t, exists)
}

func TestAwaitBucketStable_ReturnsOnNotFound(t *testing.T) {
	client := Client{get: func(context.Context, string) (api.Response, error) {
		return api.Response{}, api.APIError{StatusCode: http.StatusNotFound}
	}}
	exists, err := buckets.AwaitActiveOrNotFound(t.Context(), client, "my-bucket", time.Minute, time.Duration(0))
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestAwaitBucketStable_ErrorsOnCustomError(t *testing.T) {
	customErr := errors.New("custom error")
	client := Client{get: func(context.Context, string) (api.Response, error) {
		return api.Response{}, customErr
	}}
	exists, err := buckets.AwaitActiveOrNotFound(t.Context(), client, "my-bucket", time.Minute, time.Duration(0))
	assert.ErrorIs(t, err, customErr)
	assert.False(t, exists)
}

func TestAwaitBucketStable_ErrorsInvalidResponseData(t *testing.T) {
	client := Client{get: func(context.Context, string) (api.Response, error) {
		return api.Response{
			Data: []byte("invalid response"),
		}, nil
	}}
	exists, err := buckets.AwaitActiveOrNotFound(t.Context(), client, "my-bucket", time.Minute, time.Duration(0))
	wantErr := &json.SyntaxError{}
	assert.ErrorAs(t, err, &wantErr)
	assert.False(t, exists)
}
