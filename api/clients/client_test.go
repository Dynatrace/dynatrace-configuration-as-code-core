// @license
// Copyright 2023 Dynatrace LLC
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

package clients

import (
	"context"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/testutils"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestClient_OK(t *testing.T) {
	responses := []testutils.ResponseDef{
		{
			GET: func(t *testing.T, request *http.Request) testutils.Response {
				return testutils.Response{
					ResponseCode: http.StatusOK,
					ResponseBody: "{}",
				}
			},
		},
		{
			GET: func(t *testing.T, request *http.Request) testutils.Response {
				return testutils.Response{
					ResponseCode: http.StatusOK,
					ResponseBody: "{}",
				}
			},
		},
		{
			POST: func(t *testing.T, request *http.Request) testutils.Response {
				return testutils.Response{
					ResponseCode: http.StatusCreated,
					ResponseBody: "{}",
				}
			},
		},
		{
			PUT: func(t *testing.T, request *http.Request) testutils.Response {
				return testutils.Response{
					ResponseCode: http.StatusAccepted,
					ResponseBody: "{}",
				}
			},
		},
		{
			DELETE: func(t *testing.T, request *http.Request) testutils.Response {
				return testutils.Response{
					ResponseCode: http.StatusAccepted,
					ResponseBody: "{}",
				}
			},
		},
		{
			PATCH: func(t *testing.T, request *http.Request) testutils.Response {
				return testutils.Response{
					ResponseCode: http.StatusAccepted,
					ResponseBody: "{}",
				}
			},
		},
	}

	server := testutils.NewHTTPTestServer(t, responses)
	defer server.Close()

	cl := NewClient(rest.NewClient(server.URL(), server.Client()), "/the/path")
	resp, err := cl.Get(context.TODO(), "id", RequestOptions{})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = cl.List(context.TODO(), RequestOptions{})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = cl.Create(context.TODO(), []byte("{}"), RequestOptions{})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	resp, err = cl.Update(context.TODO(), "id", []byte("{}"), RequestOptions{})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	resp, err = cl.Delete(context.TODO(), "id", RequestOptions{})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	resp, err = cl.Patch(context.TODO(), "id", []byte("{}"), RequestOptions{})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestClient_NOK(t *testing.T) {
	responses := []testutils.ResponseDef{}

	server := testutils.NewHTTPTestServer(t, responses)
	defer server.Close()

	cl := NewClient(rest.NewClient(server.URL(), server.FaultyClient()), "/the/path")
	resp, err := cl.Get(context.TODO(), "id", RequestOptions{})
	assert.Error(t, err)
	assert.Nil(t, resp)

	resp, err = cl.List(context.TODO(), RequestOptions{})
	assert.Error(t, err)
	assert.Nil(t, resp)

	resp, err = cl.Create(context.TODO(), []byte("{}"), RequestOptions{})
	assert.Error(t, err)
	assert.Nil(t, resp)

	resp, err = cl.Update(context.TODO(), "id", []byte("{}"), RequestOptions{})
	assert.Error(t, err)
	assert.Nil(t, resp)

	resp, err = cl.Delete(context.TODO(), "id", RequestOptions{})
	assert.Error(t, err)
	assert.Nil(t, resp)

	resp, err = cl.Patch(context.TODO(), "id", []byte("{}"), RequestOptions{})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestClient_IdMissing(t *testing.T) {
	responses := []testutils.ResponseDef{}

	server := testutils.NewHTTPTestServer(t, responses)
	defer server.Close()

	cl := NewClient(rest.NewClient(server.URL(), server.Client()), "/the/path")
	resp, err := cl.Get(context.TODO(), "", RequestOptions{})
	assert.Error(t, err)
	assert.Nil(t, resp)

	resp, err = cl.Delete(context.TODO(), "", RequestOptions{})
	assert.Error(t, err)
	assert.Nil(t, resp)

	resp, err = cl.Update(context.TODO(), "", []byte("{}"), RequestOptions{})
	assert.Error(t, err)
	assert.Nil(t, resp)

	resp, err = cl.Patch(context.TODO(), "", []byte("{}"), RequestOptions{})
	assert.Error(t, err)
	assert.Nil(t, resp)
}
