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

package rest

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestIsRequest(t *testing.T) {
	// Test when RequestResponse contains an HTTP request
	req := &http.Request{}
	rr := RequestResponse{Request: req}
	request, isRequest := rr.IsRequest()

	assert.True(t, isRequest)
	assert.Equal(t, req, request)

	// Test when RequestResponse does not contain an HTTP request
	rr = RequestResponse{}
	request, isRequest = rr.IsRequest()

	assert.False(t, isRequest)
	assert.Nil(t, request)
}

func TestIsResponse(t *testing.T) {
	// Test when RequestResponse contains an HTTP response
	resp := &http.Response{}
	rr := RequestResponse{Response: resp}
	response, isResponse := rr.IsResponse()

	assert.True(t, isResponse)
	assert.Equal(t, resp, response)

	// Test when RequestResponse does not contain an HTTP response
	rr = RequestResponse{}
	response, isResponse = rr.IsResponse()

	assert.False(t, isResponse)
	assert.Nil(t, response)
}

func TestOnRequest(t *testing.T) {
	var capturedRequest RequestResponse
	callback := func(rr RequestResponse) {
		capturedRequest = rr
	}
	listener := HTTPListener{Callback: callback}

	req := &http.Request{}
	listener.onRequest("testID", req)

	assert.Equal(t, "testID", capturedRequest.ID)
	assert.Equal(t, req, capturedRequest.Request)
}

func TestOnResponse(t *testing.T) {
	// Test the onResponse method of HTTPListener
	var capturedResponse RequestResponse
	callback := func(rr RequestResponse) {
		capturedResponse = rr
	}
	listener := HTTPListener{Callback: callback}

	resp := &http.Response{}
	err := errors.New("test error")
	listener.onResponse("testID", resp, err)

	assert.Equal(t, "testID", capturedResponse.ID)
	assert.Equal(t, resp, capturedResponse.Response)
	assert.Equal(t, err, capturedResponse.Error)
}
