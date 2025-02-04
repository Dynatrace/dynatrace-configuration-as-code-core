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
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessResponse(t *testing.T) {
	t.Run("successful response with empty body", func(t *testing.T) {
		given := http.Response{StatusCode: http.StatusCreated, Body: nil}

		actual, err := ProcessResponse(&given)

		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
		assert.Nil(t, actual.Data)
	})

	t.Run("successful response when nil is provided as transformers", func(t *testing.T) {
		given := http.Response{StatusCode: http.StatusCreated}

		actual, err := ProcessResponse(&given, nil)

		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
		assert.Nil(t, actual.Data)
	})

	t.Run("return error APIError for unsuccessful response - an error is returned", func(t *testing.T) {
		given := http.Response{StatusCode: http.StatusInternalServerError}

		actual, err := ProcessResponse(&given)

		assert.Empty(t, actual)
		assert.Error(t, err)
	})

	t.Run("transform function cause an error - an error is returned", func(t *testing.T) {
		given := http.Response{StatusCode: http.StatusOK}
		transformer := func([]byte) ([]byte, error) { return nil, errors.New("an error") }

		actual, err := ProcessResponse(&given, transformer)

		assert.Empty(t, actual)
		assert.Error(t, err)
	})
}
