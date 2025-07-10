// @license
// Copyright 2024 Dynatrace LLC
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

import "fmt"

// ClientError represents a custom error type used in the client package to wrap errors returned by the rest.Client.
// It provides additional context such as the operation, resource, and reason for the error.
type ClientError struct {
	// Wrapped is the original error returned by the rest.Client.
	Wrapped error

	// Identifier is an optional string used to uniquely identify
	// the resource or entity involved in the failed operation.
	Identifier string

	// Operation is the HTTP method (e.g., GET, POST, PUT, DELETE) http/method.go
	// used in the request that caused the error.
	Operation string

	// Resource specifies the Dynatrace API resource involved in the request,
	// such as "segments", "documents", etc.
	Resource string

	// Reason provides a human-readable explanation of why the error occurred.
	Reason string
}

func (e ClientError) Error() string {
	if e.Identifier != "" {
		return fmt.Sprintf("failed to %s %s with id %s: %v", e.Operation, e.Resource, e.Identifier, e.Wrapped)
	}

	return fmt.Sprintf("failed to %s %s: %v", e.Operation, e.Resource, e.Wrapped)
}

func (e ClientError) Unwrap() error {
	return e.Wrapped
}

// ValidationError represents a client-side validation error that occurs
// during request construction. It is not an error returned from a validation
// endpoint of the API, but rather indicates that the request does not meet
// certain requirements before being sent.
type ValidationError struct {
	// Field is the name of the field that failed validation.
	Field string

	// Reason provides a human-readable explanation of why the validation failed.
	Reason string
}

func (e ValidationError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("validation failed for field %s: %s.", e.Field, e.Reason)
	}

	return fmt.Sprintf("validation failed for field %s. ", e.Field)
}

// RuntimeError represents an error that occurs when the program makes assumptions
// about the structure or content of data returned from an API, which are required
// for constructing a valid follow-up request. It wraps the original error and
// provides additional context.
type RuntimeError struct {
	// Wrapped is the original error that triggered this runtime failure.
	// It contains the low-level error details.
	Wrapped error

	// Resource is the name of the resource involved in the failed operation.
	// This helps identify which part of the system or API the error relates to.
	Resource string

	// Reason explains why the runtime assumption failed or what condition was not met.
	// This is useful for understanding the nature of the failure.
	Reason string

	// Identifier optionally provides a specific ID or key related to the resource.
	// If set, it will be included in the error message for more precise context.
	Identifier string
}

func (e RuntimeError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("request with id %s invalid: %s. %v", e.Identifier, e.Reason, e.Wrapped)
	}

	return fmt.Sprintf("request with id %s invalid: %v", e.Identifier, e.Wrapped)
}

func (e RuntimeError) Unwrap() error { return e.Wrapped }
