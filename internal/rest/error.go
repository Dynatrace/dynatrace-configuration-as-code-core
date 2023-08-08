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

import "fmt"

// HTTPError represents a custom error type containing the actual HTTP response code and payload.
type HTTPError struct {
	Code    int
	Payload []byte
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP request failed with status code %d", e.Code)
}
