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
	"encoding/json"
	"fmt"
)

// Response represents an API response
type Response struct {
	StatusCode int    `json:"-"`
	Data       []byte `json:"-"`
}

// ListResponse represents a multi-object API response
// It contains both the full JSON Data, and a slice of Objects for more convenient access
type ListResponse struct {
	Response
	Objects [][]byte `json:"-"`
}

// IsSuccess returns true if the response indicates a successful HTTP status code.
// A status code between 200 and 299 (inclusive) is considered a success.
func (r Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode <= 299
}

// Is4xxError returns true if the response indicates a 4xx client error HTTP status code.
// A status code between 400 and 499 (inclusive) is considered a client error.
func (r Response) Is4xxError() bool {
	return r.StatusCode >= 400 && r.StatusCode <= 499
}

// Is5xxError returns true if the response indicates a 5xx server error HTTP status code.
// A status code between 500 and 599 (inclusive) is considered a server error.
func (r Response) Is5xxError() bool {
	return r.StatusCode >= 500 && r.StatusCode <= 599
}

// DecodeJSON tries to unmarshal the data of the response to obj
func (r Response) DecodeJSON(obj interface{}) error {
	if err := json.Unmarshal(r.Data, obj); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// DecodeJSONObjects unmarshalls the Objects contained in the given ListResponse into a slice of T.
// To decode the full JSON response contained in ListResponse use Response.DecodeJSON.
func DecodeJSONObjects[T any](r ListResponse) ([]T, error) {
	res := make([]T, len(r.Objects))
	for i, o := range r.Objects {
		var t T
		if err := json.Unmarshal(o, &t); err != nil {
			return []T{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		res[i] = t
	}

	return res, nil
}
