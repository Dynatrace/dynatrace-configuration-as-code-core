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

package errors

import (
	"fmt"
)

// ErrorClient represents error wrapper that wraps errors happening on the clients
type ErrorClient struct {
	// Wrapped contained the raw error returned from the client
	Wrapped error

	// Identifier being not empty the error message will contain the value set
	Identifier string

	// Operation is the HTTP method that was used by the client
	Operation string

	// Resource is a string of the resource
	Resource string

	// Reason can be provided why the error is happening
	Reason string
}

func (e ErrorClient) Error() string {
	if e.Identifier != "" {
		return fmt.Sprintf("failed to %s %s with id %s: %v", e.Operation, e.Resource, e.Identifier, e.Wrapped)
	}

	return fmt.Sprintf("failed to %s %s: %v", e.Operation, e.Resource, e.Wrapped)
}

func (e ErrorClient) Unwrap() error {
	return e.Wrapped
}

type ErrorValidation struct {
	// Wrapped contained the raw error
	Wrapped error

	// Field represents the name of the field that the validation failed
	Field string

	// Reason contains why the validation failed
	Reason string
}

func (e ErrorValidation) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("validation failed for field %s: %s. %v", e.Field, e.Reason, e.Wrapped)
	}

	return fmt.Sprintf("validation failed for field %s. %v", e.Field, e.Wrapped)
}

func (e ErrorValidation) Unwrap() error {
	return e.Wrapped
}

type ErrorRuntime struct {
	// Wrapped contained the raw error
	Wrapped error

	// Resource is a string of the resource
	Resource string

	// Reason contains why the validation failed
	Reason string

	// Identifier being not empty the error message will contain the value set
	Identifier string
}

func (e ErrorRuntime) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("request with id %s invalid: %s. %v", e.Identifier, e.Reason, e.Wrapped)
	}

	return fmt.Sprintf("request with id %s invalid: %v", e.Identifier, e.Wrapped)
}

func (e ErrorRuntime) Unwrap() error { return e.Wrapped }
