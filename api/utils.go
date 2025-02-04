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

package api

import (
	"io"
	"net/http"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

// ProcessResponse processes an HTTP response and applies a transformation function to the response body if provided.
func ProcessResponse(httpResponse *http.Response, transformers ...func([]byte) ([]byte, error)) (Response, error) {
	var body []byte
	var err error

	if httpResponse.Body != nil {
		if body, err = io.ReadAll(httpResponse.Body); err != nil {
			return Response{}, NewAPIErrorFromResponseAndBody(httpResponse, body)
		}
	}

	if !rest.IsSuccess(httpResponse) {
		return Response{}, NewAPIErrorFromResponseAndBody(httpResponse, body)
	}

	if transformers != nil {
		for _, t := range transformers {
			if t != nil {
				if body, err = t(body); err != nil {
					return Response{}, NewAPIErrorFromResponseAndBody(httpResponse, body)
				}
			}
		}
	}

	return NewResponseFromHTTPResponseAndBody(httpResponse, body), nil
}
