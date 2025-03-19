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

package api_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

func TestResponse_DecodeJSON(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		data := []byte(`{"key": "value"}`)
		response := api.Response{Data: data}
		type objType map[string]string

		obj, err := api.DecodeJSON[objType](response)

		assert.NoError(t, err)
		assert.Equal(t, "value", obj["key"])
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		data := []byte(`invalid-json`)
		response := api.Response{Data: data}

		type objType map[string]string

		_, err := api.DecodeJSON[objType](response)

		assert.Error(t, err)
		assert.Equal(t, "failed to unmarshal JSON: invalid character 'i' looking for beginning of value", err.Error())
	})
}
func TestListResponse_DecodeJSON(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		data := []byte(`{"results": [ { "key": "one" }, { "key": "two" } ] }`)
		response := api.ListResponse{
			Response: api.Response{
				Data: data,
			},
		}
		type objType map[string][]map[string]string

		obj, err := api.DecodeJSON[objType](response.Response)

		assert.NoError(t, err)
		assert.Len(t, obj["results"], 2)
		assert.Equal(t, "one", obj["results"][0]["key"])
		assert.Equal(t, "two", obj["results"][1]["key"])
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		data := []byte(`invalid-json`)
		response := api.ListResponse{
			Response: api.Response{
				Data: data,
			},
		}
		type objType map[string][]map[string]string

		_, err := api.DecodeJSON[objType](response.Response)

		assert.Error(t, err)
		assert.Equal(t, "failed to unmarshal JSON: invalid character 'i' looking for beginning of value", err.Error())
	})
}

func TestDecodeJSONObjects(t *testing.T) {
	response := api.ListResponse{
		Objects: [][]byte{
			[]byte(`{ "key": "one" }`),
			[]byte(`{ "key": "two" }`),
			[]byte(`{ "key": "three" }`),
		},
	}
	type val struct {
		Key string `json:"key"`
	}

	res, err := api.DecodeJSONObjects[val](response)

	assert.NoError(t, err)
	assert.Len(t, res, 3)
	assert.Equal(t, "one", res[0].Key)
	assert.Equal(t, "two", res[1].Key)
	assert.Equal(t, "three", res[2].Key)
}

func TestDecodePaginatedJSONObjects(t *testing.T) {
	response := api.PagedListResponse{
		api.ListResponse{
			Objects: [][]byte{
				[]byte(`{ "key": "one" }`),
				[]byte(`{ "key": "two" }`),
				[]byte(`{ "key": "three" }`),
			},
		},
		api.ListResponse{
			Objects: [][]byte{
				[]byte(`{ "key": "four" }`),
			},
		},
		api.ListResponse{
			Objects: [][]byte{
				[]byte(`{ "key": "five" }`),
			},
		},
		api.ListResponse{
			Objects: [][]byte{
				[]byte(`{ "key": "six" }`),
			},
		},
	}
	type val struct {
		Key string `json:"key"`
	}

	res, err := api.DecodePaginatedJSONObjects[val](response)

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
		expected   api.APIError
		expectedOK bool
	}{
		{
			name:       "Not an API error (2xx)",
			statusCode: http.StatusOK,
			expected:   api.APIError{},
			expectedOK: false,
		},
		{
			name:       "Not an API error (3xx)",
			statusCode: http.StatusNotModified,
			expected:   api.APIError{},
			expectedOK: false,
		},
		{
			name:       "API error (4xx)",
			statusCode: http.StatusNotFound,
			expected: api.APIError{
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
			expected: api.APIError{
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
			resp := api.Response{
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
	pr := api.PagedListResponse{
		api.ListResponse{
			Response: api.Response{},
			Objects: [][]byte{
				{'1'},
				{'2'},
			},
		},
		api.ListResponse{
			Response: api.Response{},
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
		given      api.PagedListResponse
		expected   api.APIError
		expectedOK bool
	}{
		{
			"empty list is not an error",
			api.PagedListResponse{},
			api.APIError{},
			false,
		},
		{
			"single entry 4xx is an error",
			api.PagedListResponse{
				api.ListResponse{
					Response: api.Response{
						StatusCode: 403,
					},
				},
			},
			api.APIError{
				StatusCode: 403,
			},
			true,
		},
		{
			"single entry 5xx is an error",
			api.PagedListResponse{
				api.ListResponse{
					Response: api.Response{
						StatusCode: 500,
					},
				},
			},
			api.APIError{
				StatusCode: 500,
			},
			true,
		},
		{
			"several entries is not an error",
			api.PagedListResponse{
				api.ListResponse{
					Response: api.Response{
						StatusCode: 403,
					},
				},
				api.ListResponse{
					Response: api.Response{
						StatusCode: 200,
					},
				},
			},
			api.APIError{},
			false,
		},
		{
			"single entry 2xx is not an error",
			api.PagedListResponse{
				api.ListResponse{
					Response: api.Response{
						StatusCode: 201,
					},
				},
			},
			api.APIError{},
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

func TestNewResponseFromHTTPResponse(t *testing.T) {
	t.Run("http response code isn't 2xx - an APIError ", func(t *testing.T) {
		given := http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader("http message content"))}

		actual, err := api.NewResponseFromHTTPResponse(&given)

		assert.Empty(t, actual)

		assert.Error(t, err)
		var apiErr api.APIError
		assert.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
		assert.Equal(t, "http message content", string(apiErr.Body))
	})

	t.Run("http response code is 2xx - OK", func(t *testing.T) {
		given := http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("http message content"))}

		actual, err := api.NewResponseFromHTTPResponse(&given)

		require.NoError(t, err)
		require.NotEmpty(t, actual)
		assert.Equal(t, http.StatusOK, actual.StatusCode)
		assert.Equal(t, "http message content", string(actual.Data))
	})

	t.Run("read-error during reading of http body - error", func(t *testing.T) {
		given := http.Response{StatusCode: http.StatusOK, Body: &stubReaderCloser{readErr: errors.New("read error")}}

		actual, err := api.NewResponseFromHTTPResponse(&given)

		assert.Empty(t, actual)
		assert.Error(t, err)
	})

	t.Run("closer is always called", func(t *testing.T) {
		tests := []struct {
			name       string
			statusCode int
			readErr    error
		}{
			{
				name:       "response is 2xx",
				statusCode: http.StatusOK,
			}, {
				name:       "response is 4xx",
				statusCode: http.StatusNotFound,
			}, {
				name:    "error while reading",
				readErr: errors.New("read error"),
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				closer := stubReaderCloser{reader: strings.NewReader("http message content"), readErr: tc.readErr}
				given := http.Response{StatusCode: tc.statusCode, Body: &closer}

				api.NewResponseFromHTTPResponse(&given)

				assert.True(t, closer.closed)
			})
		}
	})
}

func TestAsResponseOrError(t *testing.T) {
	t.Run("Error returns error", func(t *testing.T) {
		resp, err := api.AsResponseOrError(nil, errors.New("some error"))
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "some error")
	})

	t.Run("4xx status returns APIError", func(t *testing.T) {
		const badRequest = "Bad request"
		mockBody := stubReaderCloser{reader: io.NopCloser(strings.NewReader(badRequest))}
		resp, err := api.AsResponseOrError(&http.Response{StatusCode: http.StatusBadRequest, Body: &mockBody}, nil)
		assert.Nil(t, resp)
		apiErr := api.APIError{}
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, apiErr.StatusCode, http.StatusBadRequest)
		assert.Equal(t, string(apiErr.Body), badRequest)
		assert.True(t, mockBody.closed)
	})

	t.Run("2xx status returns response", func(t *testing.T) {
		const content = "content"
		mockBody := stubReaderCloser{reader: strings.NewReader(content)}
		resp, err := api.AsResponseOrError(&http.Response{StatusCode: http.StatusOK, Body: &mockBody}, nil)
		require.NotNil(t, resp)
		assert.Equal(t, string(resp.Data), content)
		assert.Equal(t, resp.StatusCode, http.StatusOK)
		require.Nil(t, err)
		assert.True(t, mockBody.closed)
	})

	t.Run("2xx status with read error returns error", func(t *testing.T) {
		const content = "content"
		mockBody := stubReaderCloser{readErr: errors.New("read error")}
		resp, err := api.AsResponseOrError(&http.Response{StatusCode: http.StatusOK, Body: &mockBody}, nil)
		require.Nil(t, resp)
		var apiErr api.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, apiErr.StatusCode, http.StatusOK)
		assert.True(t, mockBody.closed)
	})
}

func TestIsNotFoundError(t *testing.T) {
	t.Run("404 error returns true", func(t *testing.T) {
		err := api.APIError{StatusCode: http.StatusNotFound}
		got := api.IsNotFoundError(err)
		assert.Equal(t, true, got)
	})

	t.Run("Different error returns true", func(t *testing.T) {
		err := api.APIError{StatusCode: 400}
		got := api.IsNotFoundError(err)
		assert.Equal(t, false, got)
	})

	t.Run("Not 404 api error returns false", func(t *testing.T) {
		err := api.APIError{StatusCode: http.StatusForbidden}
		got := api.IsNotFoundError(err)
		assert.Equal(t, false, got)
	})

	t.Run("Not api error returns false", func(t *testing.T) {
		err := customErr{StatusCode: http.StatusNotFound}
		got := api.IsNotFoundError(err)
		assert.Equal(t, false, got)
	})
}

type customErr struct {
	StatusCode int
}

func (e customErr) Error() string {
	return "error"
}

type stubReaderCloser struct {
	reader  io.Reader
	readErr error
	closed  bool
}

func (r *stubReaderCloser) Read(p []byte) (int, error) {
	if r.readErr != nil {
		return 0, r.readErr
	}
	return r.reader.Read(p)
}
func (r *stubReaderCloser) Close() error {
	r.closed = true
	return nil
}
