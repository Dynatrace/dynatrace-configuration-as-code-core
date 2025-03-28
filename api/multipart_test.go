/*
 * @license
 * Copyright 2024 Dynatrace LLC
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

package api_test

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/stretchr/testify/assert"
)

func TestNewMultipartResponse(t *testing.T) {
	partOne := api.Part{
		FormName: "metadata",
		FileName: "metadata.json",
		Content:  []byte(`{"key":"value"}`),
	}

	request := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   "something.com",
			Path:   "/metadata",
		},
	}

	response := http.Response{
		StatusCode: http.StatusOK,
		Request:    &request,
		Body:       io.NopCloser(strings.NewReader(`{"key":"value"}`)),
	}

	multiResponse := api.NewMultipartResponse(&response, partOne)
	assert.Equal(t, multiResponse.Request.Method, request.Method)
	assert.Equal(t, multiResponse.Request.URL, request.URL.String())
	assert.Equal(t, len(multiResponse.Parts), 1)
	assert.Equal(t, multiResponse.Parts[0], partOne)
	assert.Equal(t, multiResponse.StatusCode, response.StatusCode)
}

func TestMultipartResponse_IsSuccess(t *testing.T) {
	t.Run("invoke over null object", func(t *testing.T) {
		var mr *api.MultipartResponse

		assert.False(t, mr.IsSuccess())
	})
}

func TestMultipartResponse_GetPartWithName(t *testing.T) {
	t.Run("part found", func(t *testing.T) {
		expectedPart := api.Part{FormName: "metadata"}
		meta := api.MultipartResponse{Parts: []api.Part{expectedPart}}

		actualPart, found := meta.GetPartWithName("metadata")
		assert.True(t, found)
		assert.Equal(t, actualPart, expectedPart)
	})
	t.Run("part missing", func(t *testing.T) {
		part := api.Part{FormName: "hello"}
		meta := api.MultipartResponse{Parts: []api.Part{part}}

		actualPart, found := meta.GetPartWithName("metadata")
		assert.False(t, found)
		assert.Equal(t, actualPart, api.Part{})
	})
}
