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

package api_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api"
)

func TestProcessResponse(t *testing.T) {
	t.Run("return error APIError for unsuccessful response - an error is returned", func(t *testing.T) {
		given := http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader(""))}

		actual, err := api.ProcessResponse(&given)

		assert.Empty(t, actual)
		assert.Error(t, err)
	})

	t.Run("successful response when nil modifier is provided", func(t *testing.T) {
		given := http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}

		actual, err := api.ProcessResponse(&given, nil)

		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
		assert.Empty(t, actual.Data)
	})

	t.Run("modifier cause an error - an error is returned", func(t *testing.T) {
		given := http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}
		modifier := func(message []byte) ([]byte, error) { return nil, errors.New("an error") }

		actual, err := api.ProcessResponse(&given, modifier)

		assert.Empty(t, actual)
		assert.Error(t, err)
	})

	t.Run("use of multiple transforms works", func(t *testing.T) {
		helloModifier := func(in []byte) ([]byte, error) {
			return []byte(fmt.Sprintf("hello %s", in)), nil
		}
		closerModifier := func(name string) func([]byte) ([]byte, error) {
			return func(in []byte) ([]byte, error) {
				return []byte(fmt.Sprintf("%s, my name is %s", in, name)), nil
			}
		}

		tests := []struct {
			name      string
			modifiers []api.ModifierFunc
			expected  string
		}{
			{
				name:      "without modifier",
				modifiers: nil,
				expected:  "world",
			},
			{
				name:      "with one modifier",
				modifiers: []api.ModifierFunc{helloModifier},
				expected:  "hello world",
			},
			{
				name:      "with two modifier",
				modifiers: []api.ModifierFunc{helloModifier, helloModifier},
				expected:  "hello hello world",
			},
			{
				name:      "with closer as modifier",
				modifiers: []api.ModifierFunc{helloModifier, closerModifier("monaco")},
				expected:  "hello world, my name is monaco",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				given := http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader("world")),
				}

				actual, err := api.ProcessResponse(&given, tc.modifiers...)
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, string(actual.Data))
			})
		}
	})
}
