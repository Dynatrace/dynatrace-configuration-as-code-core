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
	"net/http"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

type Part struct {
	FormName, FileName string
	Content            []byte
}

// MultipartResponse represent an HTTP multipart response. Each field part of a multipart content is in the map indexed with the name of the part.
type MultipartResponse struct {
	StatusCode int
	Parts      []Part
	Request    rest.RequestInfo
}

func (mr *MultipartResponse) IsSuccess() bool {
	return mr != nil && mr.StatusCode >= 200 && mr.StatusCode <= 299
}

func (mr *MultipartResponse) GetPartWithName(name string) (Part, bool) {
	if mr.Parts != nil {
		for _, part := range mr.Parts {
			if part.FormName == name {
				return part, true
			}
		}
	}
	return Part{}, false
}

func NewMultipartResponse(resp *http.Response, parts ...Part) *MultipartResponse {
	return &MultipartResponse{
		StatusCode: resp.StatusCode,
		Parts:      parts,
		Request: rest.RequestInfo{
			Method: resp.Request.Method,
			URL:    resp.Request.URL.String()},
	}
}
