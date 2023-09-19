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

package api

import (
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestResponse_DecodeJSON(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		data := []byte(`{"key": "value"}`)
		response := Response{Data: data}
		type objType map[string]string

		obj, err := DecodeJSON[objType](response)

		assert.NoError(t, err)
		assert.Equal(t, "value", obj["key"])
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		data := []byte(`invalid-json`)
		response := Response{Data: data}

		type objType map[string]string

		_, err := DecodeJSON[objType](response)

		assert.Error(t, err)
		assert.Equal(t, "failed to unmarshal JSON: invalid character 'i' looking for beginning of value", err.Error())
	})
}
func TestListResponse_DecodeJSON(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		data := []byte(`{"results": [ { "key": "one" }, { "key": "two" } ] }`)
		response := ListResponse{
			Response: Response{
				Data: data,
			},
		}
		type objType map[string][]map[string]string

		obj, err := DecodeJSON[objType](response.Response)

		assert.NoError(t, err)
		assert.Len(t, obj["results"], 2)
		assert.Equal(t, "one", obj["results"][0]["key"])
		assert.Equal(t, "two", obj["results"][1]["key"])
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		data := []byte(`invalid-json`)
		response := ListResponse{
			Response: Response{
				Data: data,
			},
		}
		type objType map[string][]map[string]string

		_, err := DecodeJSON[objType](response.Response)

		assert.Error(t, err)
		assert.Equal(t, "failed to unmarshal JSON: invalid character 'i' looking for beginning of value", err.Error())
	})
}

func TestDecodeJSONObjects(t *testing.T) {
	response := ListResponse{
		Objects: [][]byte{
			[]byte(`{ "key": "one" }`),
			[]byte(`{ "key": "two" }`),
			[]byte(`{ "key": "three" }`),
		},
	}
	type val struct {
		Key string `json:"key"`
	}

	res, err := DecodeJSONObjects[val](response)

	assert.NoError(t, err)
	assert.Len(t, res, 3)
	assert.Equal(t, "one", res[0].Key)
	assert.Equal(t, "two", res[1].Key)
	assert.Equal(t, "three", res[2].Key)
}

func TestDecodePaginatedJSONObjects(t *testing.T) {
	response := PagedListResponse{
		ListResponse{
			Objects: [][]byte{
				[]byte(`{ "key": "one" }`),
				[]byte(`{ "key": "two" }`),
				[]byte(`{ "key": "three" }`),
			},
		},
		ListResponse{
			Objects: [][]byte{
				[]byte(`{ "key": "four" }`),
			},
		},
		ListResponse{
			Objects: [][]byte{
				[]byte(`{ "key": "five" }`),
			},
		},
		ListResponse{
			Objects: [][]byte{
				[]byte(`{ "key": "six" }`),
			},
		},
	}
	type val struct {
		Key string `json:"key"`
	}

	res, err := DecodePaginatedJSONObjects[val](response)

	assert.NoError(t, err)
	assert.Len(t, res, 6)
	assert.Equal(t, "one", res[0].Key)
	assert.Equal(t, "two", res[1].Key)
	assert.Equal(t, "three", res[2].Key)
	assert.Equal(t, "four", res[3].Key)
	assert.Equal(t, "five", res[4].Key)
	assert.Equal(t, "six", res[5].Key)
}

func TestAsAPIError(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		expected   APIError
		expectedOK bool
	}{
		{
			name:       "Not an API error (2xx)",
			statusCode: http.StatusOK,
			expected:   APIError{},
			expectedOK: false,
		},
		{
			name:       "Not an API error (3xx)",
			statusCode: http.StatusNotModified,
			expected:   APIError{},
			expectedOK: false,
		},
		{
			name:       "API error (4xx)",
			statusCode: http.StatusNotFound,
			expected: APIError{
				StatusCode: http.StatusNotFound,
				Request: rest.RequestInfo{
					Method: http.MethodGet,
					URL:    "https://www.dt.com/resources",
				},
			},
			expectedOK: true,
		},
		{
			name:       "API error (5xx)",
			statusCode: http.StatusServiceUnavailable,
			expected: APIError{
				StatusCode: http.StatusServiceUnavailable,
				Request: rest.RequestInfo{
					Method: http.MethodGet,
					URL:    "https://www.dt.com/resources",
				},
			},
			expectedOK: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := Response{
				StatusCode: tc.statusCode,
				Request: rest.RequestInfo{
					Method: http.MethodGet,
					URL:    "https://www.dt.com/resources",
				},
			}

			err, ok := resp.AsAPIError()

			assert.IsType(t, tc.expected, err)
			assert.Equal(t, tc.expected, err)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestPagedListResponse(t *testing.T) {
	pr := PagedListResponse{
		ListResponse{
			Response: Response{},
			Objects: [][]byte{
				{'1'},
				{'2'},
			},
		},
		ListResponse{
			Response: Response{},
			Objects: [][]byte{
				{'3'},
				{'4'},
			},
		},
	}

	assert.Equal(t, [][]byte{{'1'}, {'2'}, {'3'}, {'4'}}, pr.All())
}

func TestPagedListResponse_AsAPIError(t *testing.T) {
	testCases := []struct {
		name       string
		given      PagedListResponse
		expected   APIError
		expectedOK bool
	}{
		{
			"empty list is not an error",
			PagedListResponse{},
			APIError{},
			false,
		},
		{
			"single entry 4xx is an error",
			PagedListResponse{
				ListResponse{
					Response: Response{
						StatusCode: 403,
					},
				},
			},
			APIError{
				StatusCode: 403,
			},
			true,
		},
		{
			"single entry 5xx is an error",
			PagedListResponse{
				ListResponse{
					Response: Response{
						StatusCode: 500,
					},
				},
			},
			APIError{
				StatusCode: 500,
			},
			true,
		},
		{
			"several entries is not an error",
			PagedListResponse{
				ListResponse{
					Response: Response{
						StatusCode: 403,
					},
				},
				ListResponse{
					Response: Response{
						StatusCode: 200,
					},
				},
			},
			APIError{},
			false,
		},
		{
			"single entry 2xx is not an error",
			PagedListResponse{
				ListResponse{
					Response: Response{
						StatusCode: 201,
					},
				},
			},
			APIError{},
			false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOK := tt.given.AsAPIError()
			assert.Equal(t, tt.expected, got)
			assert.Equal(t, tt.expectedOK, gotOK)
		})
	}
}
