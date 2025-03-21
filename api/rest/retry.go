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
	"net/http"
	"time"
)

type RetryFunc func(resp *http.Response) bool

// RetryOptions represents a component for retrying failed HTTP requests.
type RetryOptions struct {
	DelayAfterRetry time.Duration
	MaxRetries      int
	ShouldRetryFunc RetryFunc
}

// RetryIfNotSuccess is a basic retry function which will retry on any non 2xx status code.
func RetryIfNotSuccess(resp *http.Response) bool {
	return !(isStatusSuccess(resp.StatusCode))
}

// RetryIfTooManyRequests return true for responses with status code Too Many Requests (429).
func RetryIfTooManyRequests(resp *http.Response) bool {
	return resp.StatusCode == http.StatusTooManyRequests
}

// RetryOnFailureExcept404 returns true for all failed responses except those with status not found.
func RetryOnFailureExcept404(resp *http.Response) bool {
	return RetryIfNotSuccess(resp) && (resp.StatusCode != http.StatusNotFound)
}

// isStatusSuccess returns true if it is a 2xx status code
func isStatusSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode <= 299
}

// ShouldRetry returns true if a retry should happen (e.g., not on 403 or 200, but on 429, etc.) which is valid for all kind of APIs
func ShouldRetry(statusCode int) bool {
	if isStatusSuccess(statusCode) {
		return false
	}

	switch statusCode {
	case http.StatusForbidden:
		return false
	default:
		return true
	}
}
