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

// IsSuccess returns true if the response indicates a successful HTTP status code.
// A status code between 200 and 299 (inclusive) is considered a success.
func (resp Response) IsSuccess() bool {
	return resp.StatusCode >= 200 && resp.StatusCode <= 299
}

// Is4xxError returns true if the response indicates a 4xx client error HTTP status code.
// A status code between 400 and 499 (inclusive) is considered a client error.
func (resp Response) Is4xxError() bool {
	return resp.StatusCode >= 400 && resp.StatusCode <= 499
}

// Is5xxError returns true if the response indicates a 5xx server error HTTP status code.
// A status code between 500 and 599 (inclusive) is considered a server error.
func (resp Response) Is5xxError() bool {
	return resp.StatusCode >= 500 && resp.StatusCode <= 599
}

// DecodeJSON tries to unmarshal the data of the response to obj
func (r Response) DecodeJSON(obj interface{}) error {
	if err := json.Unmarshal(r.Data, obj); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}
