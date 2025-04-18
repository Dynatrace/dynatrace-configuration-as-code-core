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

package rest_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dynatrace/dynatrace-configuration-as-code-core/api/rest"
)

func TestShouldRetry(t *testing.T) {
	t.Run("Should be false if status is 200", func(t *testing.T) {
		got := rest.ShouldRetry(http.StatusOK)
		assert.False(t, got)
	})

	t.Run("Should be false if status is 403", func(t *testing.T) {
		got := rest.ShouldRetry(http.StatusForbidden)
		assert.False(t, got)
	})

	t.Run("Should be true if status is not 403 or 2xx", func(t *testing.T) {
		got := rest.ShouldRetry(http.StatusNotFound)
		assert.True(t, got)
	})
}
