/*
 * @license
 * Copyright 2024 Dynatrace LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
	"github.com/stretchr/testify/assert"
)

func TestClientError_Error(t *testing.T) {
	t.Run("without identifier", func(t *testing.T) {
		expectedErrorMessage := "failed to GET segments: test error"
		err := api.ClientError{
			Wrapped:   errors.New("test error"),
			Operation: http.MethodGet,
			Resource:  "segments",
			Reason:    "this is a reason",
		}

		assert.Equal(t, expectedErrorMessage, err.Error())
	})
	t.Run("with identifier", func(t *testing.T) {
		expectedErrorMessage := "failed to GET segments with id 123: test error"
		err := api.ClientError{
			Wrapped:    errors.New("test error"),
			Operation:  http.MethodGet,
			Resource:   "segments",
			Reason:     "this is a reason",
			Identifier: "123",
		}

		assert.Equal(t, expectedErrorMessage, err.Error())
	})
}

func TestClientError_Unwrap(t *testing.T) {
	expectedErr := errors.New("test error")
	err := api.ClientError{
		Wrapped:   expectedErr,
		Operation: http.MethodGet,
		Resource:  "segments",
		Reason:    "this is a reason",
	}

	assert.Equal(t, expectedErr, err.Unwrap())
}

func TestValidationError_Error(t *testing.T) {
	t.Run("without reason", func(t *testing.T) {
		expectedErrorMessage := "validation failed for field field."
		err := api.ValidationError{
			Field: "field",
		}

		assert.Equal(t, expectedErrorMessage, err.Error())
	})
	t.Run("with reason", func(t *testing.T) {
		expectedErrorMessage := "validation failed for field field: this is a reason."
		err := api.ValidationError{
			Field:  "field",
			Reason: "this is a reason",
		}

		assert.Equal(t, expectedErrorMessage, err.Error())
	})
}

func TestRuntimeError_Error(t *testing.T) {
	t.Run("without reason", func(t *testing.T) {
		expectedErrorMessage := "request with id 123 invalid: test error"
		err := api.RuntimeError{
			Wrapped:    errors.New("test error"),
			Resource:   "resource",
			Identifier: "123",
		}

		assert.Equal(t, expectedErrorMessage, err.Error())
	})
	t.Run("with reason", func(t *testing.T) {
		expectedErrorMessage := "request with id 123 invalid: this is a reason. test error"
		err := api.RuntimeError{
			Wrapped:    errors.New("test error"),
			Resource:   "resource",
			Reason:     "this is a reason",
			Identifier: "123",
		}

		assert.Equal(t, expectedErrorMessage, err.Error())
	})
}

func TestRuntimeError_Unwrap(t *testing.T) {
	expectedErr := errors.New("test error")
	err := api.RuntimeError{
		Wrapped:    expectedErr,
		Resource:   "segments",
		Reason:     "this is a reason",
		Identifier: "123",
	}

	assert.Equal(t, expectedErr, err.Unwrap())
}
